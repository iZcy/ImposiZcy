package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/freetype/truetype"
	"github.com/iZcy/imposizcy/internal/models"
	"github.com/iZcy/imposizcy/internal/repositories"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

// FontRegistry caches parsed truetype fonts keyed by family.
// It looks up unknown families through the FontRepository, reads the file from disk,
// and falls back to a baked-in Go font when nothing matches.
type FontRegistry struct {
	repo      *repositories.FontRepository
	uploadDir string
	logger    *logrus.Logger

	mu       sync.RWMutex
	cache    map[string]*truetype.Font // family -> regular
	cacheB   map[string]*truetype.Font // family -> bold (if uploaded separately as "<family> Bold")
	fallback *truetype.Font
	bold     *truetype.Font
}

func NewFontRegistry(repo *repositories.FontRepository, uploadDir string, logger *logrus.Logger) (*FontRegistry, error) {
	reg, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return nil, fmt.Errorf("parse fallback regular font: %w", err)
	}
	bold, err := truetype.Parse(gobold.TTF)
	if err != nil {
		return nil, fmt.Errorf("parse fallback bold font: %w", err)
	}
	return &FontRegistry{
		repo:      repo,
		uploadDir: uploadDir,
		logger:    logger,
		cache:     map[string]*truetype.Font{},
		cacheB:    map[string]*truetype.Font{},
		fallback:  reg,
		bold:      bold,
	}, nil
}

// Face returns a font.Face for the requested family/size/bold combo.
// Resolution order: in-memory cache → repository lookup (and parse) → built-in fallback.
func (r *FontRegistry) Face(ctx context.Context, family string, size float64, bold bool) font.Face {
	tt := r.resolve(ctx, family, bold)
	return truetype.NewFace(tt, &truetype.Options{Size: size, DPI: 72, Hinting: font.HintingFull})
}

func (r *FontRegistry) resolve(ctx context.Context, family string, bold bool) *truetype.Font {
	if family == "" {
		if bold {
			return r.bold
		}
		return r.fallback
	}

	r.mu.RLock()
	if bold {
		if f, ok := r.cacheB[family]; ok {
			r.mu.RUnlock()
			return f
		}
	}
	if f, ok := r.cache[family]; ok {
		r.mu.RUnlock()
		// If bold requested but no bold variant cached, return regular; gg may synthesize.
		if bold {
			return f
		}
		return f
	}
	r.mu.RUnlock()

	tt, err := r.loadFromRepo(ctx, family)
	if err != nil || tt == nil {
		if err != nil && r.logger != nil {
			r.logger.WithError(err).WithField("family", family).Debug("font lookup failed, using fallback")
		}
		if bold {
			return r.bold
		}
		return r.fallback
	}

	r.mu.Lock()
	r.cache[family] = tt
	r.mu.Unlock()
	return tt
}

func (r *FontRegistry) loadFromRepo(ctx context.Context, family string) (*truetype.Font, error) {
	if r.repo == nil {
		return nil, nil
	}
	row, err := r.repo.GetByFamily(ctx, family, "")
	if err != nil || row == nil {
		return nil, err
	}
	return r.parseFile(row)
}

func (r *FontRegistry) parseFile(row *models.Font) (*truetype.Font, error) {
	abs := row.Path
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(r.uploadDir, row.Path)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("read font %q: %w", abs, err)
	}
	tt, err := truetype.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse font %q: %w", abs, err)
	}
	return tt, nil
}

// Register parses and caches a freshly-uploaded font immediately so the first render after upload is hot.
func (r *FontRegistry) Register(row *models.Font) error {
	tt, err := r.parseFile(row)
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.cache[row.Family] = tt
	r.mu.Unlock()
	return nil
}

// Invalidate drops a family from the cache (call after deletion).
func (r *FontRegistry) Invalidate(family string) {
	r.mu.Lock()
	delete(r.cache, family)
	delete(r.cacheB, family)
	r.mu.Unlock()
}
