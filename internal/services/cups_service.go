package services

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/iZcy/imposizcy/internal/models"
	"github.com/sirupsen/logrus"
)

type CUPSPrinterInfo struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	IsDefault bool  `json:"is_default"`
}

type CUPSJobInfo struct {
	CUPSJobID string `json:"cups_job_id"`
	Printer   string `json:"printer"`
	Status    string `json:"status"`
	Size      string `json:"size,omitempty"`
}

type CUPSService struct {
	logger        *logrus.Logger
	pollInterval  time.Duration
	pollTimeout   time.Duration
}

func NewCUPSService(logger *logrus.Logger) *CUPSService {
	return &CUPSService{
		logger:       logger,
		pollInterval: 5 * time.Second,
		pollTimeout:  5 * time.Minute,
	}
}

func (s *CUPSService) Available() bool {
	_, err := exec.LookPath("lpstat")
	if err != nil {
		s.logger.Warn("lpstat not found in PATH — CUPS unavailable")
		return false
	}
	cmd := exec.Command("lpstat", "-r")
	if out, err := cmd.Output(); err != nil || !strings.Contains(string(out), "running") {
		s.logger.WithError(err).Warn("CUPS scheduler not running")
		return false
	}
	return true
}

func (s *CUPSService) ListPrinters() ([]CUPSPrinterInfo, error) {
	cmd := exec.Command("lpstat", "-p", "-d")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("lpstat -p -d: %w", err)
	}

	var printers []CUPSPrinterInfo
	var defaultPrinter string

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "system default destination:") {
			defaultPrinter = strings.TrimPrefix(line, "system default destination:")
			defaultPrinter = strings.TrimSpace(defaultPrinter)
			continue
		}

		if !strings.HasPrefix(line, "printer ") {
			continue
		}

		name, status := parseLpstatPrinterLine(line)
		if name == "" {
			continue
		}

		printers = append(printers, CUPSPrinterInfo{
			Name:      name,
			Status:    string(mapCUPSPrinterStatus(status)),
			IsDefault: name == defaultPrinter,
		})
	}

	if len(printers) == 0 && defaultPrinter != "" {
		printers = append(printers, CUPSPrinterInfo{
			Name:      defaultPrinter,
			Status:    "idle",
			IsDefault: true,
		})
	}

	return printers, nil
}

type SubmitJobOptions struct {
	CUPSName  string
	FilePath  string
	Copies    int
	PaperSize string
	ColorMode string
	Sides     string
}

func (s *CUPSService) SubmitJob(opts SubmitJobOptions) (string, error) {
	args := []string{
		"-d", opts.CUPSName,
	}

	if opts.Copies > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", opts.Copies))
	}
	if opts.PaperSize != "" {
		args = append(args, "-o", "media="+opts.PaperSize)
	}
	if opts.ColorMode != "" {
		switch opts.ColorMode {
		case "bw", "grayscale", "mono":
			args = append(args, "-o", "ColorModel=Gray", "-o", "print-color-mode=monochrome")
		case "color", "colour":
			args = append(args, "-o", "ColorModel=CMYK", "-o", "print-color-mode=color")
		}
	}
	if opts.Sides != "" {
		switch opts.Sides {
		case "two-sided", "duplex", "double-sided":
			args = append(args, "-o", "sides=two-sided-long-edge")
		case "one-sided", "simplex", "single-sided":
			args = append(args, "-o", "sides=one-sided")
		}
	}

	args = append(args, opts.FilePath)

	s.logger.WithFields(logrus.Fields{
		"printer":    opts.CUPSName,
		"file":       opts.FilePath,
		"copies":     opts.Copies,
		"paper_size": opts.PaperSize,
		"color_mode": opts.ColorMode,
	}).Info("Submitting print job to CUPS")

	cmd := exec.Command("lp", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("lp command failed: %w\noutput: %s", err, string(out))
	}

	cupsJobID := parseLPJobID(string(out))
	if cupsJobID == "" {
		return "", fmt.Errorf("could not parse CUPS job ID from lp output: %s", string(out))
	}

	s.logger.WithField("cups_job_id", cupsJobID).Info("CUPS job submitted")
	return cupsJobID, nil
}

func (s *CUPSService) GetPrinterStatus(cupsName string) (models.PrinterStatus, error) {
	cmd := exec.Command("lpstat", "-p", cupsName)
	out, err := cmd.Output()
	if err != nil {
		return models.PrinterStatusOffline, fmt.Errorf("lpstat -p %s: %w", cupsName, err)
	}

	_, status := parseLpstatPrinterLine(string(out))
	return mapCUPSPrinterStatus(status), nil
}

func (s *CUPSService) GetJobStatus(cupsJobID string) (string, error) {
	cmd := exec.Command("lpstat", "-o")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("lpstat -o: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, cupsJobID+" ") || strings.HasPrefix(line, cupsJobID+"\t") {
			if strings.Contains(strings.ToLower(line), "held") {
				return "held", nil
			}
			return "printing", nil
		}
	}

	return "done", nil
}

func (s *CUPSService) PollJobCompletion(ctx context.Context, cupsJobID string) error {
	ctx, cancel := context.WithTimeout(ctx, s.pollTimeout)
	defer cancel()

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("print job %s polling timed out", cupsJobID)
		case <-ticker.C:
			status, err := s.GetJobStatus(cupsJobID)
			if err != nil {
				s.logger.WithError(err).Warn("Error polling CUPS job status")
				continue
			}
			switch status {
			case "done":
				return nil
			case "held":
				return fmt.Errorf("print job %s is held in CUPS queue", cupsJobID)
			case "printing":
				s.logger.WithFields(logrus.Fields{
					"cups_job_id": cupsJobID,
					"status":      status,
				}).Debug("Print job still in progress")
			}
		}
	}
}

func (s *CUPSService) CancelJob(cupsJobID string) error {
	cmd := exec.Command("cancel", cupsJobID)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cancel %s: %w\noutput: %s", cupsJobID, err, string(out))
	}
	return nil
}

var lpJobIDRe = regexp.MustCompile(`request id is (\S+)`)

func parseLPJobID(output string) string {
	m := lpJobIDRe.FindStringSubmatch(output)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

var lpstatPrinterRe = regexp.MustCompile(`^printer\s+(\S+)\s+is\s+(\S+)`)

func parseLpstatPrinterLine(line string) (name, status string) {
	m := lpstatPrinterRe.FindStringSubmatch(line)
	if len(m) < 3 {
		return "", ""
	}
	name = m[1]
	status = m[2]
	status = strings.TrimRight(status, ".")
	return name, status
}

func mapCUPSPrinterStatus(status string) models.PrinterStatus {
	switch strings.ToLower(status) {
	case "idle":
		return models.PrinterStatusIdle
	case "printing", "now", "processing":
		return models.PrinterStatusPrinting
	case "disabled", "offline", "stopped":
		return models.PrinterStatusOffline
	default:
		return models.PrinterStatusError
	}
}
