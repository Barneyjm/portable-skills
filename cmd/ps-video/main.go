package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Barneyjm/portable-skills/pkg/video"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "ps-video",
	Short: "Video metadata skill",
	Long:  "Read metadata from video files. No external dependencies required.",
}

func init() {
	rootCmd.AddCommand(newInfoCmd())
	rootCmd.AddCommand(newFormatsCmd())
}

func newInfoCmd() *cobra.Command {
	var jsonOutput bool
	var detailed bool

	cmd := &cobra.Command{
		Use:   "info <file> [file...]",
		Short: "Display video metadata",
		Long:  "Read and display metadata from video files including duration, resolution, codecs, and more.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var results []interface{}

			for _, path := range args {
				if !video.IsSupportedFormat(path) {
					if jsonOutput {
						results = append(results, map[string]interface{}{
							"file":  path,
							"error": fmt.Sprintf("unsupported format: %s", path),
						})
						continue
					}
					fmt.Fprintf(os.Stderr, "Warning: Unsupported format: %s\n", path)
					continue
				}

				if detailed {
					info, err := video.GetDetailedInfo(path)
					if err != nil {
						if jsonOutput {
							results = append(results, map[string]interface{}{
								"file":  path,
								"error": err.Error(),
							})
							continue
						}
						fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
						continue
					}

					if jsonOutput {
						results = append(results, info)
					} else {
						printDetailedInfo(info)
						if len(args) > 1 {
							fmt.Println()
						}
					}
				} else {
					meta, err := video.Info(path)
					if err != nil {
						if jsonOutput {
							results = append(results, map[string]interface{}{
								"file":  path,
								"error": err.Error(),
							})
							continue
						}
						fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
						continue
					}

					if jsonOutput {
						results = append(results, meta)
					} else {
						printMetadata(meta)
						if len(args) > 1 {
							fmt.Println()
						}
					}
				}
			}

			if jsonOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if len(results) == 1 {
					return encoder.Encode(results[0])
				}
				return encoder.Encode(results)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed track information")

	return cmd
}

func newFormatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "formats",
		Short: "List supported video formats",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Supported video formats:")
			for _, f := range video.SupportedFormats() {
				fmt.Printf("  - %s\n", f)
			}
			return nil
		},
	}
}

func printMetadata(m *video.Metadata) {
	fmt.Printf("File: %s\n", m.File)
	fmt.Printf("Format: %s\n", m.Format)

	if m.Duration > 0 {
		fmt.Printf("Duration: %s (%.2f seconds)\n", m.DurationStr, m.Duration)
	}

	if m.Resolution != "" {
		fmt.Printf("Resolution: %s\n", m.Resolution)
	}

	if m.VideoCodec != "" {
		fmt.Printf("Video Codec: %s\n", m.VideoCodec)
	}

	if m.FrameRate > 0 {
		fmt.Printf("Frame Rate: %.2f fps\n", m.FrameRate)
	}

	if m.AudioCodec != "" {
		fmt.Printf("Audio Codec: %s\n", m.AudioCodec)
		if m.AudioChannels > 0 {
			channelStr := "mono"
			if m.AudioChannels == 2 {
				channelStr = "stereo"
			} else if m.AudioChannels > 2 {
				channelStr = fmt.Sprintf("%d channels", m.AudioChannels)
			}
			fmt.Printf("Audio Channels: %s\n", channelStr)
		}
		if m.AudioSampleRate > 0 {
			fmt.Printf("Audio Sample Rate: %d Hz\n", m.AudioSampleRate)
		}
	}

	if m.TotalBitrate > 0 {
		fmt.Printf("Bitrate: %d kbps\n", m.TotalBitrate)
	}

	if m.CreationTime != nil {
		fmt.Printf("Created: %s\n", m.CreationTime.Format("2006-01-02 15:04:05"))
	}

	if m.ModificationTime != nil && (m.CreationTime == nil || !m.CreationTime.Equal(*m.ModificationTime)) {
		fmt.Printf("Modified: %s\n", m.ModificationTime.Format("2006-01-02 15:04:05"))
	}

	fmt.Printf("Size: %s\n", formatSize(m.SizeBytes))
}

func printDetailedInfo(info *video.DetailedInfo) {
	printMetadata(&info.Metadata)

	if len(info.Tracks) > 0 {
		fmt.Println("\nTracks:")
		for _, track := range info.Tracks {
			fmt.Printf("  Track %d: %s", track.ID, track.Type)
			if track.Codec != "" {
				fmt.Printf(" (%s)", track.Codec)
			}
			if track.Width > 0 && track.Height > 0 {
				fmt.Printf(" %dx%d", track.Width, track.Height)
			}
			if track.Channels > 0 {
				fmt.Printf(" %dch", track.Channels)
			}
			if track.SampleRate > 0 {
				fmt.Printf(" %dHz", track.SampleRate)
			}
			fmt.Println()
		}
	}
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
