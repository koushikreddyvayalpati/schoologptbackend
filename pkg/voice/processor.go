package voice

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	speech "cloud.google.com/go/speech/apiv1"
	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"github.com/schoolgpt/backend/internal/config"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

// Processor handles voice processing operations
type Processor struct {
	config             *config.Config
	speechClient       *speech.Client
	textToSpeechClient *texttospeech.Client
}

// New creates a new voice processor
func New(cfg *config.Config) (*Processor, error) {
	ctx := context.Background()

	// Create Speech-to-Text client
	speechClient, err := speech.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating speech client: %v", err)
	}

	// Create Text-to-Speech client
	textToSpeechClient, err := texttospeech.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating text-to-speech client: %v", err)
	}

	return &Processor{
		config:             cfg,
		speechClient:       speechClient,
		textToSpeechClient: textToSpeechClient,
	}, nil
}

// Close closes the voice processor clients
func (p *Processor) Close() error {
	err1 := p.speechClient.Close()
	err2 := p.textToSpeechClient.Close()

	if err1 != nil {
		return err1
	}
	return err2
}

// TranscribeAudio transcribes audio to text
func (p *Processor) TranscribeAudio(ctx context.Context, audioData []byte, encoding AudioEncoding, sampleRateHertz int32, languageCode string) (string, error) {
	// Configure the request
	req := &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:        encoding.ToSpeechEncoding(),
			SampleRateHertz: sampleRateHertz,
			LanguageCode:    languageCode,
			Model:           "default", // Can be "default", "command_and_search", "phone_call", etc.
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{
				Content: audioData,
			},
		},
	}

	// Send the request
	resp, err := p.speechClient.Recognize(ctx, req)
	if err != nil {
		return "", fmt.Errorf("error recognizing speech: %v", err)
	}

	// Process the response
	var transcript string
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			transcript += alt.Transcript
		}
	}

	return transcript, nil
}

// SynthesizeSpeech converts text to speech
func (p *Processor) SynthesizeSpeech(ctx context.Context, text, outputPath string, voiceName string, languageCode string) error {
	// Configure the request
	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{
				Text: text,
			},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: languageCode,
			Name:         voiceName,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	// Send the request
	resp, err := p.textToSpeechClient.SynthesizeSpeech(ctx, req)
	if err != nil {
		return fmt.Errorf("error synthesizing speech: %v", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	// Write the audio content to file
	if err := os.WriteFile(outputPath, resp.AudioContent, 0644); err != nil {
		return fmt.Errorf("error writing audio file: %v", err)
	}

	return nil
}

// ProcessVoiceQuery processes a voice query and returns a voice response
func (p *Processor) ProcessVoiceQuery(ctx context.Context, audioFile string, outputFile string) (string, error) {
	// Read audio file
	audioData, err := os.ReadFile(audioFile)
	if err != nil {
		return "", fmt.Errorf("error reading audio file: %v", err)
	}

	// Detect file format and encoding
	encoding, sampleRate, err := detectAudioFormat(audioFile)
	if err != nil {
		return "", err
	}

	// Transcribe audio
	transcript, err := p.TranscribeAudio(ctx, audioData, encoding, sampleRate, "en-US")
	if err != nil {
		return "", err
	}

	// Process the transcript (in a real app, this would call GPT or another service)
	response := fmt.Sprintf("I heard you say: %s", transcript)

	// Synthesize speech
	err = p.SynthesizeSpeech(ctx, response, outputFile, "en-US-Wavenet-D", "en-US")
	if err != nil {
		return "", err
	}

	return transcript, nil
}

// detectAudioFormat is a helper function to detect audio format
// In a real application, this would analyze the file to determine its format
func detectAudioFormat(filename string) (AudioEncoding, int32, error) {
	ext := filepath.Ext(filename)

	switch ext {
	case ".wav":
		return AudioEncodingLinear16, 16000, nil
	case ".flac":
		return AudioEncodingFlac, 16000, nil
	case ".ogg":
		return AudioEncodingOggOpus, 16000, nil
	default:
		return AudioEncodingUnspecified, 0,
			fmt.Errorf("unsupported audio format: %s", ext)
	}
}
