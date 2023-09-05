package test_utils

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/faiface/beep"
	"github.com/mjibson/go-dsp/fft"
	"math"
	"testing"
)

// GetSimilarity Get a similarity score between two audio files
func GetSimilarity(a1, a2 [][2]float64) float64 {

	// Extract and compare each channel of the stereo audio pair
	fA1Left := make([]float64, len(a1))
	fA1Right := make([]float64, len(a1))

	for i, sample := range a1 {
		fA1Left[i] = extractSpectralFeatures([]float64{sample[0]})[0]
		fA1Right[i] = extractSpectralFeatures([]float64{sample[1]})[0]
	}

	fA2Left := make([]float64, len(a2))
	fA2Right := make([]float64, len(a2))

	for i, sample := range a2 {
		fA2Left[i] = extractSpectralFeatures([]float64{sample[0]})[0]
		fA2Right[i] = extractSpectralFeatures([]float64{sample[1]})[0]
	}

	// Both score are 0..1, we're going to average them
	jcLeft := jaccardSimilarity(fA1Left, fA2Left)
	jcRight := jaccardSimilarity(fA1Right, fA2Right)
	return (jcLeft + jcRight) / 2
}

// GetAllSamples Get all samples from a stream
func GetSamples(t *testing.T, s beep.Streamer, bufferSize int) [][2]float64 {
	samples := make([][2]float64, bufferSize)
	_, ok := s.Stream(samples[:])
	if !ok {
		t.Fatal("Could not get any data from stream")
	}
	return samples
}

// GetAllSamples Get all samples from a stream
// Beware that live stream will not contains any length information
// So this method won't work properly
func GetAllSamples(t *testing.T, s beep.StreamSeekCloser) [][2]float64 {
	return GetSamples(t, s, s.Len())
}

// Calculate Jaccard similarity between the two tracks
// @see https://en.wikipedia.org/wiki/Jaccard_index
func jaccardSimilarity(trackA, trackB []float64) float64 {
	tA := mapset.NewSet[float64]()
	tB := mapset.NewSet[float64]()
	for _, featureA := range trackA {
		tA.Add(featureA)
	}

	for _, featureB := range trackB {
		tB.Add(featureB)
	}

	union := float64(tA.Union(tB).Cardinality())
	intersect := float64(tA.Intersect(tB).Cardinality())
	return intersect / union
}

// Uses FFT to extract main audio features
func extractSpectralFeatures(samples []float64) []float64 {
	// Perform FFT on the samples
	spectrum := fft.FFTReal(samples)

	// Calculate magnitude of complex values in the spectrum
	maxLen := int(math.Max(1, float64(len(spectrum))/2.0))
	magnitudes := make([]float64, maxLen)
	for i, complexValue := range spectrum[:maxLen] {
		magnitudes[i] = math.Sqrt(real(complexValue)*real(complexValue) + imag(complexValue)*imag(complexValue))
	}

	return magnitudes
}
