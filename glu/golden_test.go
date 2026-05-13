package main_test

import (
	"bytes"
	"flag"
	"fmt"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Golden-PDF regression tests.
//
// Two-stage compare:
//
//  1. Render input → PDF with --source-date-epoch fixed. Compare
//     bytes against expected.pdf. Match → test passes silently.
//  2. Bytes differ → render both PDFs to PNG via pdftoppm, compare
//     pixel-by-pixel with a small AA tolerance. If pixel diff is
//     below the threshold the test still passes but t.Logf reports
//     the byte drift (signal: "structurally different, visually
//     identical"). Above threshold → t.Errorf, plus actual.pdf and
//     actual.png are written next to the golden for inspection.
//
// Run with -update to regenerate expected.pdf and expected.png from
// the current glu output:
//
//	go test ./glu -run Golden -update
//
// Fixtures live in testdata/golden/<name>/ with input.md (or
// input.html) and the two expected files. Single-page only — multi-
// page support is left for later.

var update = flag.Bool("update", false, "regenerate golden expected.pdf / expected.png")

// fixedEpoch makes the PDF CreationDate / XMP timestamps / UUIDs
// reproducible across runs. The actual seconds value is arbitrary;
// what matters is that it stays the same.
const fixedEpoch = "1700000000"

// pixelDiffTolerance: per-channel absolute difference below this is
// considered "same pixel". The 8-bit equivalent threshold here is
// ~3/256 — wide enough to absorb sub-pixel AA jitter between
// pdftoppm versions, narrow enough to catch real layout shifts.
const pixelDiffChannelTolerance = 256 * 3

// pixelDiffFailRatio: fraction of differing pixels above which the
// test fails. 0.5% is a balance between "twitchy" and "useless".
const pixelDiffFailRatio = 0.005

// gluBinary is the path to the built glu test binary. Set by
// TestMain so each subtest can exec it directly without re-running
// `go build`.
var gluBinary string

func TestMain(m *testing.M) {
	flag.Parse()
	tmp, err := os.MkdirTemp("", "glu-golden-bin-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create tempdir:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmp)
	bin := filepath.Join(tmp, "glu")
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build glu test binary: %v\n%s", err, out)
		os.Exit(1)
	}
	gluBinary = bin
	os.Exit(m.Run())
}

func TestGolden(t *testing.T) {
	entries, err := filepath.Glob("../testdata/golden/*")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Skip("no golden fixtures under testdata/golden/ — populate and re-run")
	}
	for _, dir := range entries {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		t.Run(filepath.Base(dir), func(t *testing.T) {
			runGoldenFixture(t, dir)
		})
	}
}

func runGoldenFixture(t *testing.T, dir string) {
	t.Helper()
	input, err := findFixtureInput(dir)
	if err != nil {
		t.Fatalf("%s: %v", dir, err)
	}

	tmp := t.TempDir()
	outPDF := filepath.Join(tmp, "out.pdf")
	cmd := exec.Command(gluBinary,
		"--source-date-epoch="+fixedEpoch,
		"--log-file=-",
		"-q",
		"-o", outPDF,
		input,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("glu failed for %s: %v\n%s", input, err, out)
	}
	actualPDF, err := os.ReadFile(outPDF)
	if err != nil {
		t.Fatal(err)
	}

	expectedPDFPath := filepath.Join(dir, "expected.pdf")
	expectedPNGPath := filepath.Join(dir, "expected.png")

	if *update {
		if err := os.WriteFile(expectedPDFPath, actualPDF, 0644); err != nil {
			t.Fatal(err)
		}
		pngBytes, err := renderPDFToPNG(t, actualPDF)
		if err != nil {
			t.Fatalf("rendering PNG: %v", err)
		}
		if err := os.WriteFile(expectedPNGPath, pngBytes, 0644); err != nil {
			t.Fatal(err)
		}
		t.Logf("updated %s and %s", expectedPDFPath, expectedPNGPath)
		return
	}

	expectedPDF, err := os.ReadFile(expectedPDFPath)
	if err != nil {
		t.Fatalf("missing %s — run with -update to create it", expectedPDFPath)
	}

	if bytes.Equal(actualPDF, expectedPDF) {
		return
	}

	// Stage 2: bytes differ → visual check.
	t.Logf("PDF bytes differ from golden (actual=%d bytes, expected=%d) — running pixel diff",
		len(actualPDF), len(expectedPDF))

	if _, err := exec.LookPath("pdftoppm"); err != nil {
		t.Skipf("pdftoppm not on PATH — install poppler to enable visual diff; byte diff alone cannot decide regression")
	}

	actualPNG, err := renderPDFToPNG(t, actualPDF)
	if err != nil {
		t.Fatalf("rendering actual PNG: %v", err)
	}
	expectedPNG, err := os.ReadFile(expectedPNGPath)
	if err != nil {
		// Drop the actual PDF for inspection and bail.
		writeActuals(t, dir, actualPDF, actualPNG)
		t.Fatalf("missing %s — actual.pdf and actual.png saved for inspection; run with -update if the new output is intended", expectedPNGPath)
	}

	ratio, err := pixelDiffRatio(actualPNG, expectedPNG)
	if err != nil {
		writeActuals(t, dir, actualPDF, actualPNG)
		t.Fatalf("pixel comparison failed: %v (actual.pdf/actual.png saved)", err)
	}
	pct := ratio * 100
	if ratio > pixelDiffFailRatio {
		writeActuals(t, dir, actualPDF, actualPNG)
		t.Errorf("visual regression: %.4f%% of pixels differ (threshold %.2f%%). actual.pdf and actual.png saved next to the golden.",
			pct, pixelDiffFailRatio*100)
		return
	}
	t.Logf("PDF byte-different but visually within tolerance (%.4f%% pixel diff, threshold %.2f%%)",
		pct, pixelDiffFailRatio*100)
}

func findFixtureInput(dir string) (string, error) {
	for _, name := range []string{"input.md", "input.html", "input.htm", "input.lua"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no input.md / input.html / input.lua in %s", dir)
}

// renderPDFToPNG runs pdftoppm at 100 DPI on a single-page PDF and
// returns the rendered PNG. The PDF is fed via a tempfile (not
// stdin, because not all pdftoppm builds accept "-" as input) and
// the output is read back from <prefix>-1.png. Multi-page PDFs are
// not supported — the test fixtures are deliberately single-page.
func renderPDFToPNG(t *testing.T, pdfBytes []byte) ([]byte, error) {
	t.Helper()
	tmp := t.TempDir()
	pdfPath := filepath.Join(tmp, "in.pdf")
	if err := os.WriteFile(pdfPath, pdfBytes, 0644); err != nil {
		return nil, err
	}
	prefix := filepath.Join(tmp, "page")
	cmd := exec.Command("pdftoppm", "-png", "-r", "100", pdfPath, prefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftoppm: %w\n%s", err, out)
	}
	// pdftoppm writes <prefix>-1.png (or -01.png with more pages).
	candidates := []string{prefix + "-1.png", prefix + "-01.png"}
	for _, c := range candidates {
		if data, err := os.ReadFile(c); err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("pdftoppm did not produce %s-{1,01}.png", prefix)
}

// pixelDiffRatio returns the fraction of pixels where any RGB
// channel differs by more than pixelDiffChannelTolerance.
// Alpha is ignored — pdftoppm output is always opaque.
func pixelDiffRatio(a, b []byte) (float64, error) {
	imgA, err := png.Decode(bytes.NewReader(a))
	if err != nil {
		return 0, fmt.Errorf("decoding actual PNG: %w", err)
	}
	imgB, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return 0, fmt.Errorf("decoding expected PNG: %w", err)
	}
	ba, bb := imgA.Bounds(), imgB.Bounds()
	if ba != bb {
		return 1.0, fmt.Errorf("image dimensions differ: actual=%v expected=%v", ba, bb)
	}
	var diff, total int
	for y := ba.Min.Y; y < ba.Max.Y; y++ {
		for x := ba.Min.X; x < ba.Max.X; x++ {
			ar, ag, ab, _ := imgA.At(x, y).RGBA()
			br, bg, bb, _ := imgB.At(x, y).RGBA()
			if absDiff(ar, br) > pixelDiffChannelTolerance ||
				absDiff(ag, bg) > pixelDiffChannelTolerance ||
				absDiff(ab, bb) > pixelDiffChannelTolerance {
				diff++
			}
			total++
		}
	}
	if total == 0 {
		return 0, fmt.Errorf("empty image")
	}
	return float64(diff) / float64(total), nil
}

func absDiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

// writeActuals dumps the diverging artefacts next to the golden so
// the developer can open them in a viewer.
func writeActuals(t *testing.T, dir string, actualPDF, actualPNG []byte) {
	t.Helper()
	_ = os.WriteFile(filepath.Join(dir, "actual.pdf"), actualPDF, 0644)
	if actualPNG != nil {
		_ = os.WriteFile(filepath.Join(dir, "actual.png"), actualPNG, 0644)
	}
}

