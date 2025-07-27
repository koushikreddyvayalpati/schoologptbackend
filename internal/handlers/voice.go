package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/schoolgpt/backend/pkg/voice"
)

// VoiceHandler handles voice-related requests
type VoiceHandler struct {
	processor *voice.Processor
}

// NewVoiceHandler creates a new voice handler
func NewVoiceHandler(processor *voice.Processor) *VoiceHandler {
	return &VoiceHandler{
		processor: processor,
	}
}

// TranscribeRequest represents a request to transcribe audio
type TranscribeRequest struct {
	LanguageCode string `form:"language_code" binding:"omitempty"`
}

// SynthesizeRequest represents a request to synthesize speech
type SynthesizeRequest struct {
	Text         string `json:"text" binding:"required"`
	LanguageCode string `json:"language_code" binding:"omitempty"`
	VoiceName    string `json:"voice_name" binding:"omitempty"`
}

// TranscribeResponse represents a response from transcription
type TranscribeResponse struct {
	Transcript string `json:"transcript"`
}

// SynthesizeResponse represents a response from synthesis
type SynthesizeResponse struct {
	AudioURL string `json:"audio_url"`
}

// Transcribe handles a request to transcribe audio
func (h *VoiceHandler) Transcribe(c *gin.Context) {
	var req TranscribeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Set default language code if not provided
	languageCode := req.LanguageCode
	if languageCode == "" {
		languageCode = "en-US"
	}

	// Get the audio file from the request
	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio file required"})
		return
	}
	defer file.Close()

	// Create a temporary directory for the audio file
	tempDir := filepath.Join(os.TempDir(), "schoolgpt", "audio")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating temporary directory"})
		return
	}

	// Generate a unique filename
	filename := fmt.Sprintf("%s_%s", userID.(string), uuid.New().String())
	audioPath := filepath.Join(tempDir, filename+filepath.Ext(header.Filename))

	// Save the audio file
	out, err := os.Create(audioPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving audio file"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error copying audio file"})
		return
	}

	// Detect audio format and encoding
	encoding, sampleRate, err := detectAudioFormat(audioPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Read the audio file
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading audio file"})
		return
	}

	// Transcribe the audio
	transcript, err := h.processor.TranscribeAudio(c.Request.Context(), audioData, encoding, sampleRate, languageCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error transcribing audio: %v", err)})
		return
	}

	// Clean up the audio file
	os.Remove(audioPath)

	// Return the transcript
	c.JSON(http.StatusOK, TranscribeResponse{
		Transcript: transcript,
	})
}

// Synthesize handles a request to synthesize speech
func (h *VoiceHandler) Synthesize(c *gin.Context) {
	var req SynthesizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Set default values if not provided
	languageCode := req.LanguageCode
	if languageCode == "" {
		languageCode = "en-US"
	}

	voiceName := req.VoiceName
	if voiceName == "" {
		voiceName = "en-US-Wavenet-D"
	}

	// Create a directory for the audio files
	audioDir := filepath.Join("static", "audio")
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating audio directory"})
		return
	}

	// Generate a unique filename
	filename := fmt.Sprintf("%s_%d_%s", userID.(string), time.Now().Unix(), uuid.New().String())
	outputPath := filepath.Join(audioDir, filename+".mp3")

	// Synthesize the speech
	err := h.processor.SynthesizeSpeech(c.Request.Context(), req.Text, outputPath, voiceName, languageCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error synthesizing speech: %v", err)})
		return
	}

	// Return the audio URL
	audioURL := fmt.Sprintf("/static/audio/%s.mp3", filename)
	c.JSON(http.StatusOK, SynthesizeResponse{
		AudioURL: audioURL,
	})
}

// detectAudioFormat is a helper function to detect audio format
func detectAudioFormat(filename string) (voice.AudioEncoding, int32, error) {
	ext := filepath.Ext(filename)

	switch ext {
	case ".wav":
		return voice.AudioEncodingLinear16, 16000, nil
	case ".flac":
		return voice.AudioEncodingFlac, 16000, nil
	case ".ogg":
		return voice.AudioEncodingOggOpus, 16000, nil
	default:
		return 0, 0, fmt.Errorf("unsupported audio format: %s", ext)
	}
}
