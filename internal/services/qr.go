package services

type qrEncoder struct {
	version  int
	size     int
	modules  [][]bool
	isFunc   [][]bool
	ecLevel  int
}

const (
	ecL = 0
	ecM = 1
	ecQ = 2
	ecH = 3
)

type versionInfo struct {
	totalCW      int
	ecCWPerBlock int
	numBlocks    int
	dataCWPerBlock int
}

var versionTable = map[int]map[int]versionInfo{
	1: {ecL: {26, 7, 1, 19}, ecM: {26, 10, 1, 16}, ecQ: {26, 13, 1, 13}, ecH: {26, 17, 1, 9}},
	2: {ecL: {44, 10, 1, 34}, ecM: {44, 16, 1, 28}, ecQ: {44, 22, 1, 22}, ecH: {44, 28, 1, 16}},
	3: {ecL: {70, 15, 1, 55}, ecM: {70, 26, 1, 44}, ecQ: {70, 18, 2, 17}, ecH: {70, 22, 2, 13}},
	4: {ecL: {100, 20, 1, 80}, ecM: {100, 18, 2, 32}, ecQ: {100, 26, 2, 24}, ecH: {100, 16, 4, 9}},
	5: {ecL: {134, 26, 1, 108}, ecM: {134, 24, 2, 43}, ecQ: {134, 18, 2, 15}, ecH: {134, 22, 2, 11}},
	6: {ecL: {172, 18, 2, 68}, ecM: {172, 16, 4, 27}, ecQ: {172, 24, 4, 19}, ecH: {172, 28, 4, 15}},
	7: {ecL: {196, 20, 2, 78}, ecM: {196, 18, 4, 31}, ecQ: {196, 18, 2, 14}, ecH: {196, 26, 4, 13}},
	8: {ecL: {242, 24, 2, 97}, ecM: {242, 22, 2, 38}, ecQ: {242, 22, 4, 18}, ecH: {242, 26, 4, 14}},
	9: {ecL: {292, 30, 2, 116}, ecM: {292, 22, 3, 36}, ecQ: {292, 20, 4, 18}, ecH: {292, 24, 4, 12}},
	10: {ecL: {346, 18, 2, 68}, ecM: {346, 26, 4, 43}, ecQ: {346, 24, 6, 19}, ecH: {346, 28, 6, 15}},
}

var alignmentPositions = map[int][]int{
	1: {}, 2: {6, 18}, 3: {6, 22}, 4: {6, 26}, 5: {6, 30},
	6: {6, 34}, 7: {6, 22, 38}, 8: {6, 24, 42}, 9: {6, 26, 46}, 10: {6, 28, 50},
}

func newQREncoder(version, ecLvl int) *qrEncoder {
	size := version*4 + 17
	modules := make([][]bool, size)
	isFunc := make([][]bool, size)
	for i := range modules {
		modules[i] = make([]bool, size)
		isFunc[i] = make([]bool, size)
	}
	return &qrEncoder{version: version, size: size, modules: modules, isFunc: isFunc, ecLevel: ecLvl}
}

func (q *qrEncoder) addFinderPattern(row, col int) {
	for r := -1; r <= 7; r++ {
		for c := -1; c <= 7; c++ {
			rr, cc := row+r, col+c
			if rr < 0 || rr >= q.size || cc < 0 || cc >= q.size {
				continue
			}
			q.isFunc[rr][cc] = true
			val := (0 <= r && r <= 6 && (c == 0 || c == 6)) ||
				(0 <= c && c <= 6 && (r == 0 || r == 6)) ||
				(2 <= r && r <= 4 && 2 <= c && c <= 4)
			q.modules[rr][cc] = val
		}
	}
}

func (q *qrEncoder) addAlignmentPatterns() {
	positions := alignmentPositions[q.version]
	for _, r := range positions {
		for _, c := range positions {
			if q.isFunc[r][c] {
				continue
			}
			for dr := -2; dr <= 2; dr++ {
				for dc := -2; dc <= 2; dc++ {
					rr, cc := r+dr, c+dc
					q.isFunc[rr][cc] = true
					val := abs(dr) == 2 || abs(dc) == 2 || (dr == 0 && dc == 0)
					q.modules[rr][cc] = val
				}
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (q *qrEncoder) addTimingPatterns() {
	for i := 8; i < q.size-8; i++ {
		if !q.isFunc[6][i] {
			q.isFunc[6][i] = true
			q.modules[6][i] = i%2 == 0
		}
		if !q.isFunc[i][6] {
			q.isFunc[i][6] = true
			q.modules[i][6] = i%2 == 0
		}
	}
}

func (q *qrEncoder) addFormatInfo(mask int) {
	formatBits := getFormatBits(q.ecLevel, mask)
	for i := 0; i < 6; i++ {
		q.setFuncModule(8, i, (formatBits>>i)&1 == 1)
	}
	q.setFuncModule(8, 7, (formatBits>>6)&1 == 1)
	q.setFuncModule(8, 8, (formatBits>>7)&1 == 1)
	q.setFuncModule(7, 8, (formatBits>>8)&1 == 1)
	for i := 9; i < 15; i++ {
		q.setFuncModule(14-i, 8, (formatBits>>i)&1 == 1)
	}
	for i := 0; i < 8; i++ {
		q.setFuncModule(q.size-1-i, 8, (formatBits>>i)&1 == 1)
	}
	for i := 8; i < 15; i++ {
		q.setFuncModule(8, q.size-15+i, (formatBits>>i)&1 == 1)
	}
	q.setFuncModule(8, q.size-8, true)
}

func (q *qrEncoder) addVersionInfo() {
	if q.version < 7 {
		return
	}
	rem := q.version
	for i := 0; i < 18; i++ {
		bit := (rem >> 17) & 1
		rem = (rem << 1) & 0x3FFFF
		if bit != 0 {
			rem ^= 0x1F25
		}
		bits := (q.version << 12) | rem
		for j := 0; j < 6; j++ {
			for k := 0; k < 3; k++ {
				val := (bits>>(j*3+k))&1 != 0
				q.setFuncModule(5-j, q.size-11+k, val)
				q.setFuncModule(q.size-11+k, 5-j, val)
			}
		}
	}
}

func (q *qrEncoder) setFuncModule(r, c int, val bool) {
	q.isFunc[r][c] = true
	q.modules[r][c] = val
}

func (q *qrEncoder) drawCodewords(data []byte) {
	i := 0
	for right := q.size - 1; right >= 1; right -= 2 {
		if right == 6 {
			right = 5
		}
		for vert := 0; vert < q.size; vert++ {
			for j := 0; j < 2; j++ {
				x := right - j
				upward := ((right+1)&2) == 0
				var y int
				if upward {
					y = q.size - 1 - vert
				} else {
					y = vert
				}
				if !q.isFunc[y][x] && i < len(data)*8 {
					q.modules[y][x] = (data[i>>3]>>(7-(i&7)))&1 == 1
					i++
				}
			}
		}
	}
}

func (q *qrEncoder) applyMask(mask int) {
	for r := 0; r < q.size; r++ {
		for c := 0; c < q.size; c++ {
			if q.isFunc[r][c] {
				continue
			}
			var invert bool
			switch mask {
			case 0:
				invert = (r+c)%2 == 0
			case 1:
				invert = r%2 == 0
			case 2:
				invert = c%3 == 0
			case 3:
				invert = (r+c)%3 == 0
			case 4:
				invert = (r/2+c/3)%2 == 0
			case 5:
				invert = (r*c)%2+(r*c)%3 == 0
			case 6:
				invert = ((r*c)%2+(r*c)%3)%2 == 0
			case 7:
				invert = ((r+c)%2+(r*c)%3)%2 == 0
			}
			if invert {
				q.modules[r][c] = !q.modules[r][c]
			}
		}
	}
}

func (q *qrEncoder) penaltyScore() int {
	score := 0
	for r := 0; r < q.size; r++ {
		runColor := false
		runLen := 0
		runHistory := make([]int, 7)
		for c := 0; c < q.size; c++ {
			if q.modules[r][c] == runColor {
				runLen++
				if runLen == 5 {
					score += 3
				} else if runLen > 5 {
					score++
				}
			} else {
				q.addFinderPenalty(runHistory, runLen, &score)
				runColor = q.modules[r][c]
				runLen = 1
			}
		}
		q.addFinderPenalty(runHistory, runLen, &score)
	}
	for c := 0; c < q.size; c++ {
		runColor := false
		runLen := 0
		runHistory := make([]int, 7)
		for r := 0; r < q.size; r++ {
			if q.modules[r][c] == runColor {
				runLen++
				if runLen == 5 {
					score += 3
				} else if runLen > 5 {
					score++
				}
			} else {
				q.addFinderPenalty(runHistory, runLen, &score)
				runColor = q.modules[r][c]
				runLen = 1
			}
		}
		q.addFinderPenalty(runHistory, runLen, &score)
	}
	for r := 0; r < q.size-1; r++ {
		for c := 0; c < q.size-1; c++ {
			v := q.modules[r][c]
			if v == q.modules[r][c+1] && v == q.modules[r+1][c] && v == q.modules[r+1][c+1] {
				score += 3
			}
		}
	}
	dark := 0
	for r := 0; r < q.size; r++ {
		for c := 0; c < q.size; c++ {
			if q.modules[r][c] {
				dark++
			}
		}
	}
	total := q.size * q.size
	ratio := dark * 100 / total
	prev := abs(ratio/5*5 - 50) / 5
	next := abs((ratio/5+1)*5 - 50) / 5
	if prev < next {
		score += prev * 10
	} else {
		score += next * 10
	}
	return score
}

func (q *qrEncoder) addFinderPenalty(history []int, curRun int, score *int) {
	history = append(history[:0], curRun)
	for i := len(history) - 1; i > 0; i-- {
		history[i] = history[i-1]
	}
	history[0] = curRun
	if len(history) >= 7 && history[1] == 1 && history[3] == 1 && history[5] == 1 &&
		(history[0]+history[2]+history[4]+history[6])%2 == 0 {
		*score += 40
	}
}

func getFormatBits(ecLvl, mask int) int {
	data := ecLvl<<3 | mask
	rem := data
	for i := 0; i < 10; i++ {
		rem = (rem << 1) & 0x7FF
		if (rem>>10)&1 != 0 {
			rem ^= 0x537
		}
	}
	bits := ((data << 10) | rem) ^ 0x5412
	return bits & 0x7FFF
}

func selectVersion(dataLen int, ecLvl int) int {
	for v := 1; v <= 10; v++ {
		info := versionTable[v][ecLvl]
		numDataCW := info.numBlocks * info.dataCWPerBlock
		modeBits := 4
		if v < 10 {
			modeBits = 4
		}
		charCountBits := 8
		if v >= 10 {
			charCountBits = 16
		}
		availableBits := numDataCW*8 - modeBits - charCountBits - 4
		if dataLen*8 <= availableBits {
			return v
		}
	}
	return 10
}

func encodeData(text string, ecLvl int) ([]byte, error) {
	data := []byte(text)
	version := selectVersion(len(data), ecLvl)
	info := versionTable[version][ecLvl]

	numDataCW := info.numBlocks * info.dataCWPerBlock

	charCountBits := 8
	if version >= 10 {
		charCountBits = 16
	}

	bitStream := make([]bool, 0, numDataCW*8)
	bitStream = appendBits(bitStream, 0b0100, 4)
	bitStream = appendBits(bitStream, len(data), charCountBits)
	for _, b := range data {
		bitStream = appendBits(bitStream, int(b), 8)
	}
	bitStream = appendBits(bitStream, 0, min(4, numDataCW*8-len(bitStream)))
	if len(bitStream)%8 != 0 {
		bitStream = appendBits(bitStream, 0, 8-len(bitStream)%8)
	}

	dataBytes := bitsToBytes(bitStream)
	for len(dataBytes) < numDataCW {
		if len(dataBytes)%2 == 0 {
			dataBytes = append(dataBytes, 0xEC)
		} else {
			dataBytes = append(dataBytes, 0x11)
		}
	}

	blocks := make([][]byte, info.numBlocks)
	idx := 0
	for i := 0; i < info.numBlocks; i++ {
		blocks[i] = make([]byte, info.dataCWPerBlock)
		copy(blocks[i], dataBytes[idx:idx+info.dataCWPerBlock])
		idx += info.dataCWPerBlock
	}

	ecBlocks := make([][]byte, info.numBlocks)
	for i, block := range blocks {
		ecBlocks[i] = reedSolomonComputeRemainder(block, info.ecCWPerBlock)
	}

	result := make([]byte, 0, info.totalCW)
	for i := 0; i < info.dataCWPerBlock; i++ {
		for j := 0; j < info.numBlocks; j++ {
			if i < len(blocks[j]) {
				result = append(result, blocks[j][i])
			}
		}
	}
	for i := 0; i < info.ecCWPerBlock; i++ {
		for j := 0; j < info.numBlocks; j++ {
			if i < len(ecBlocks[j]) {
				result = append(result, ecBlocks[j][i])
			}
		}
	}

	return result, nil
}

func appendBits(stream []bool, val, len int) []bool {
	for i := len - 1; i >= 0; i-- {
		stream = append(stream, (val>>i)&1 == 1)
	}
	return stream
}

func bitsToBytes(bits []bool) []byte {
	result := make([]byte, (len(bits)+7)/8)
	for i, b := range bits {
		if b {
			result[i/8] |= 1 << (7 - i%8)
		}
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func reedSolomonComputeRemainder(data []byte, degree int) []byte {
	result := make([]byte, degree)
	for _, b := range data {
		factor := int(b) ^ int(result[0])
		copy(result, result[1:])
		result[len(result)-1] = 0
		if factor != 0 {
			for i := 0; i < len(result); i++ {
				result[i] = byte(int(result[i]) ^ rsMul(generatorPoly(degree)[i+1], factor))
			}
		}
	}
	return result
}

var generatorCache = map[int][]int{}

func generatorPoly(degree int) []int {
	if p, ok := generatorCache[degree]; ok {
		return p
	}
	roots := make([]int, degree)
	for i := 0; i < degree; i++ {
		roots[i] = i
	}
	poly := []int{1}
	for _, root := range roots {
		newPoly := make([]int, len(poly)+1)
		for i, c := range poly {
			newPoly[i] ^= rsMul(c, 1)
			newPoly[i+1] ^= rsMul(c, gexp(root))
		}
		poly = newPoly
	}
	generatorCache[degree] = poly
	return poly
}

var expTable = make([]int, 512)
var logTable = make([]int, 256)

func init() {
	x := 1
	for i := 0; i < 255; i++ {
		expTable[i] = x
		logTable[x] = i
		x = (x << 1) ^ x
		if x >= 256 {
			x ^= 0x11D
		}
	}
	for i := 255; i < 512; i++ {
		expTable[i] = expTable[i-255]
	}
}

func gexp(a int) int {
	return expTable[a%255]
}

func rsMul(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	return expTable[logTable[a]+logTable[b]]
}

func generateQR(text string) ([]bool, int, error) {
	ecLvl := ecM

	data, err := encodeData(text, ecLvl)
	if err != nil {
		return nil, 0, err
	}

	version := selectVersion(len([]byte(text)), ecLvl)

	q := newQREncoder(version, ecLvl)

	q.addFinderPattern(0, 0)
	q.addFinderPattern(0, q.size-7)
	q.addFinderPattern(q.size-7, 0)
	q.addAlignmentPatterns()
	q.addTimingPatterns()
	q.addFormatInfo(0)
	q.addVersionInfo()

	modulesCopy := make([][]bool, q.size)
	for i := range modulesCopy {
		modulesCopy[i] = make([]bool, q.size)
	}

	bestMask := 0
	bestScore := int(^uint(0) >> 1)

	for mask := 0; mask < 8; mask++ {
		for r := 0; r < q.size; r++ {
			copy(modulesCopy[r], q.modules[r])
		}
		q.applyMask(mask)
		q.addFormatInfo(mask)
		s := q.penaltyScore()
		if s < bestScore {
			bestScore = s
			bestMask = mask
		}
		for r := 0; r < q.size; r++ {
			copy(q.modules[r], modulesCopy[r])
		}
		for r := 0; r < q.size; r++ {
			for c := 0; c < q.size; c++ {
				if !q.isFunc[r][c] {
					q.modules[r][c] = false
				}
			}
		}
	}

	q.drawCodewords(data)
	q.applyMask(bestMask)
	q.addFormatInfo(bestMask)

	flat := make([]bool, q.size*q.size)
	for r := 0; r < q.size; r++ {
		for c := 0; c < q.size; c++ {
			flat[r*q.size+c] = q.modules[r][c]
		}
	}
	return flat, q.size, nil
}
