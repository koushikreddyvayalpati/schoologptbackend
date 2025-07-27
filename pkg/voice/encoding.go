package voice

import (
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

// AudioEncoding represents the encoding of an audio file
type AudioEncoding int

const (
	// AudioEncodingUnspecified default value
	AudioEncodingUnspecified AudioEncoding = iota
	// AudioEncodingLinear16 uncompressed 16-bit signed little-endian samples (Linear PCM)
	AudioEncodingLinear16
	// AudioEncodingFlac FLAC (Free Lossless Audio Codec)
	AudioEncodingFlac
	// AudioEncodingOggOpus OGG_OPUS (Opus codec in OGG container)
	AudioEncodingOggOpus
)

// ToSpeechEncoding converts an AudioEncoding to a speechpb.RecognitionConfig_AudioEncoding
func (e AudioEncoding) ToSpeechEncoding() speechpb.RecognitionConfig_AudioEncoding {
	switch e {
	case AudioEncodingLinear16:
		return speechpb.RecognitionConfig_LINEAR16
	case AudioEncodingFlac:
		return speechpb.RecognitionConfig_FLAC
	case AudioEncodingOggOpus:
		return speechpb.RecognitionConfig_OGG_OPUS
	default:
		return speechpb.RecognitionConfig_ENCODING_UNSPECIFIED
	}
}
