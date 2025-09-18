package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/arthurfung/project-github-cronjob/pkg/provider"
	"gopkg.in/yaml.v3"
)

type ChannelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type MessageInfo struct {
	Channel   string    `json:"channel"`
	ChannelID string    `json:"channel_id"`
	User      string    `json:"user"`
	Text      string    `json:"text"`
	Timestamp string    `json:"timestamp"`
	Date      time.Time `json:"date"`
}

// Configuration structures
type Config struct {
	Storage    StorageConfig    `yaml:"storage"`
	Slack      SlackConfig      `yaml:"slack"`
	Defaults   DefaultsConfig   `yaml:"defaults"`
	FileNaming FileNamingConfig `yaml:"file_naming"`
}

type StorageConfig struct {
	Type string   `yaml:"type"`
	B2   B2Config `yaml:"b2"`
}

type B2Config struct {
	BucketName       string `yaml:"bucket_name"`
	ApplicationKeyID string `yaml:"application_key_id"`
	ApplicationKey   string `yaml:"application_key"`
	PathPrefix       string `yaml:"path_prefix"`
	Endpoint         string `yaml:"endpoint"`
}

type SlackConfig struct {
	XOXCToken string `yaml:"xoxc_token"`
	XOXDToken string `yaml:"xoxd_token"`
	XOXPToken string `yaml:"xoxp_token"`
}

type DefaultsConfig struct {
	LimitPerChannel int    `yaml:"limit_per_channel"`
	OutputFormat    string `yaml:"output_format"`
	ChannelsFilter  string `yaml:"channels_filter"`
}

type FileNamingConfig struct {
	IncludeTimestamp bool   `yaml:"include_timestamp"`
	IncludeDateRange bool   `yaml:"include_date_range"`
	IncludeChannels  bool   `yaml:"include_channels"`
	Prefix           string `yaml:"prefix"`
}

func main() {
	var (
		startDate  = flag.String("start", "", "Start date (YYYY-MM-DD format)")
		endDate    = flag.String("end", "", "End date (YYYY-MM-DD format)")
		limit      = flag.Int("limit", 0, "Maximum messages per channel (0 = use config default)")
		output     = flag.String("output", "", "Output format: console, json, csv (empty = use config default)")
		channels   = flag.String("channels", "", "Channels to fetch: all, public, private, or comma-separated list (empty = use config default)")
		configFile = flag.String("config", "config.yaml", "Path to configuration file")
	)
	flag.Parse()

	// Load configuration
	cfg, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Apply defaults from config if flags not provided
	if *limit == 0 {
		*limit = cfg.Defaults.LimitPerChannel
	}
	if *output == "" {
		*output = cfg.Defaults.OutputFormat
	}
	if *channels == "" {
		*channels = cfg.Defaults.ChannelsFilter
	}

	// Parse date range
	var oldest, latest string
	if *startDate != "" {
		startTime, err := time.Parse("2006-01-02", *startDate)
		if err != nil {
			log.Fatalf("Invalid start date format: %v", err)
		}
		oldest = fmt.Sprintf("%.6f", float64(startTime.Unix()))
	}

	if *endDate != "" {
		endTime, err := time.Parse("2006-01-02", *endDate)
		if err != nil {
			log.Fatalf("Invalid end date format: %v", err)
		}
		// Add 24 hours to include the entire end date
		endTime = endTime.Add(24 * time.Hour)
		latest = fmt.Sprintf("%.6f", float64(endTime.Unix()))
	}

	// Initialize API provider
	apiProvider := provider.New()
	err := apiProvider.RefreshUsers(context.Background())
	if err != nil {
		log.Fatalf("Failed to refresh users: %v", err)
	}
	err = apiProvider.RefreshChannels(context.Background())
	if err != nil {
		log.Fatalf("Failed to refresh channels: %v", err)
	}

	log.Printf("Fetching conversation history...")
	log.Printf("Date range: %s to %s", *startDate, *endDate)
	log.Printf("Limit per channel: %d", *limit)
	log.Printf("Channels filter: %s", *channels)

	// Get all channels
	channelTypes := []string{}
	switch *channels {
	case "all":
		channelTypes = []string{"public_channel", "private_channel", "mpim", "im"}
	case "public":
		channelTypes = []string{"public_channel"}
	case "private":
		channelTypes = []string{"private_channel"}
	default:
		// Specific channel names provided
		if *channels != "" {
			specificChannels := strings.Split(*channels, ",")
			err := fetchSpecificChannels(apiProvider, specificChannels, oldest, latest, *limit, *output, cfg, *startDate, *endDate, *channels)
			if err != nil {
				log.Fatalf("Failed to fetch specific channels: %v", err)
			}
			return
		}
	}

	// Fetch channels
	log.Printf("Fetching channels of types: %v", channelTypes)
	channels_list := apiProvider.GetChannels(context.Background(), channelTypes)

	log.Printf("Found %d channels", len(channels_list))

	var allMessages []MessageInfo

	// Fetch conversation history for each channel
	for i, channel := range channels_list {
		log.Printf("Fetching history for channel %d/%d: %s (%s)", i+1, len(channels_list), channel.Name, channel.ID)

		messages, err := apiProvider.GetConversationHistory(context.Background(), channel.ID, *limit, oldest, latest, "", false)
		if err != nil {
			log.Printf("Warning: Failed to fetch history for channel %s: %v", channel.Name, err)
			continue
		}

		log.Printf("  Found %d messages in %s", len(messages), channel.Name)

		// Convert messages to our format
		for _, msg := range messages {
			timestamp, _ := strconv.ParseFloat(msg.Timestamp, 64)
			msgTime := time.Unix(int64(timestamp), 0)

			messageInfo := MessageInfo{
				Channel:   channel.Name,
				ChannelID: channel.ID,
				User:      msg.UserID,
				Text:      msg.Text,
				Timestamp: msg.Timestamp,
				Date:      msgTime,
			}
			allMessages = append(allMessages, messageInfo)
		}
	}

	log.Printf("Total messages collected: %d", len(allMessages))

	// Output results
	err = outputMessages(allMessages, *output, cfg, *startDate, *endDate, *channels)
	if err != nil {
		log.Fatalf("Failed to output messages: %v", err)
	}
}

func fetchSpecificChannels(apiProvider *provider.ApiProvider, channelNames []string, oldest, latest string, limit int, output string, cfg *Config, startDate, endDate, channels string) error {
	var allMessages []MessageInfo

	for i, channelName := range channelNames {
		channelName = strings.TrimSpace(channelName)
		log.Printf("Fetching history for specific channel %d/%d: %s", i+1, len(channelNames), channelName)

		messages, err := apiProvider.GetConversationHistory(context.Background(), channelName, limit, oldest, latest, "", false)
		if err != nil {
			log.Printf("Warning: Failed to fetch history for channel %s: %v", channelName, err)
			continue
		}

		log.Printf("  Found %d messages in %s", len(messages), channelName)

		// Convert messages to our format
		for _, msg := range messages {
			timestamp, _ := strconv.ParseFloat(msg.Timestamp, 64)
			msgTime := time.Unix(int64(timestamp), 0)

			messageInfo := MessageInfo{
				Channel:   channelName,
				ChannelID: channelName, // For specific channels, we use the name as ID
				User:      msg.UserID,
				Text:      msg.Text,
				Timestamp: msg.Timestamp,
				Date:      msgTime,
			}
			allMessages = append(allMessages, messageInfo)
		}
	}

	log.Printf("Total messages collected from specific channels: %d", len(allMessages))
	return outputMessages(allMessages, output, cfg, startDate, endDate, channels)
}

// loadConfig loads and parses the YAML configuration file with environment variable expansion
func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	expandedData := expandEnvVars(string(data))

	var cfg Config
	err = yaml.Unmarshal([]byte(expandedData), &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// expandEnvVars expands environment variables in the format ${VAR_NAME}
func expandEnvVars(input string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name from ${VAR_NAME}
		varName := match[2 : len(match)-1]
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match // Return original if env var not found
	})
}

// generateFilename creates a filename based on the configuration and parameters
func generateFilename(cfg *Config, format, startDate, endDate, channels string) string {
	parts := []string{}

	if cfg.FileNaming.Prefix != "" {
		parts = append(parts, cfg.FileNaming.Prefix)
	}

	if cfg.FileNaming.IncludeTimestamp {
		timestamp := time.Now().Format("20060102-150405")
		parts = append(parts, timestamp)
	}

	if cfg.FileNaming.IncludeDateRange && (startDate != "" || endDate != "") {
		dateRange := ""
		if startDate != "" && endDate != "" {
			dateRange = fmt.Sprintf("%s-to-%s", startDate, endDate)
		} else if startDate != "" {
			dateRange = fmt.Sprintf("from-%s", startDate)
		} else if endDate != "" {
			dateRange = fmt.Sprintf("until-%s", endDate)
		}
		if dateRange != "" {
			parts = append(parts, dateRange)
		}
	}

	if cfg.FileNaming.IncludeChannels && channels != "" && channels != "all" {
		// Sanitize channel names for filename
		sanitized := strings.ReplaceAll(channels, ",", "-")
		sanitized = strings.ReplaceAll(sanitized, " ", "")
		parts = append(parts, sanitized)
	}

	filename := strings.Join(parts, "_")
	if filename == "" {
		filename = "conversations"
	}

	return filename + "." + format
}

// uploadToB2 uploads data to B2 storage using S3-compatible API
func uploadToB2(cfg *Config, data []byte, filename string) error {
	if cfg.Storage.Type != "b2" {
		return nil // Skip if not B2 storage
	}

	// Configure AWS SDK for B2
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.Storage.B2.ApplicationKeyID,
			cfg.Storage.B2.ApplicationKey,
			"",
		)),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client for B2
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Storage.B2.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Storage.B2.Endpoint)
		} else {
			// Default B2 S3-compatible endpoint
			o.BaseEndpoint = aws.String("https://s3.us-west-000.backblazeb2.com")
		}
		o.UsePathStyle = true
	})

	// Construct the full key path
	key := filename
	if cfg.Storage.B2.PathPrefix != "" {
		key = strings.TrimSuffix(cfg.Storage.B2.PathPrefix, "/") + "/" + filename
	}

	// Upload to B2
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(cfg.Storage.B2.BucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})

	if err != nil {
		return fmt.Errorf("failed to upload to B2: %w", err)
	}

	log.Printf("Successfully uploaded to B2: %s/%s", cfg.Storage.B2.BucketName, key)
	return nil
}

func outputMessages(messages []MessageInfo, format string, cfg *Config, startDate, endDate, channels string) error {
	switch format {
	case "json":
		// Always output to console
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		err := encoder.Encode(messages)
		if err != nil {
			return err
		}

		// Upload to B2 if configured
		if cfg != nil && cfg.Storage.Type == "b2" {
			jsonData, err := json.MarshalIndent(messages, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON for B2 upload: %w", err)
			}
			filename := generateFilename(cfg, "json", startDate, endDate, channels)
			err = uploadToB2(cfg, jsonData, filename)
			if err != nil {
				log.Printf("Warning: Failed to upload to B2: %v", err)
			}
		}

	case "csv":
		// Always output to console
		fmt.Println("Channel,ChannelID,User,Timestamp,Date,Text")
		for _, msg := range messages {
			// Escape quotes in text for CSV
			text := strings.ReplaceAll(msg.Text, "\"", "\"\"")
			fmt.Printf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n",
				msg.Channel, msg.ChannelID, msg.User, msg.Timestamp,
				msg.Date.Format("2006-01-02 15:04:05"), text)
		}

		// Upload to B2 if configured
		if cfg != nil && cfg.Storage.Type == "b2" {
			var csvBuffer bytes.Buffer
			writer := csv.NewWriter(&csvBuffer)

			// Write header
			writer.Write([]string{"Channel", "ChannelID", "User", "Timestamp", "Date", "Text"})

			// Write data
			for _, msg := range messages {
				record := []string{
					msg.Channel,
					msg.ChannelID,
					msg.User,
					msg.Timestamp,
					msg.Date.Format("2006-01-02 15:04:05"),
					msg.Text,
				}
				writer.Write(record)
			}
			writer.Flush()

			if err := writer.Error(); err != nil {
				return fmt.Errorf("failed to generate CSV for B2 upload: %w", err)
			}

			filename := generateFilename(cfg, "csv", startDate, endDate, channels)
			err := uploadToB2(cfg, csvBuffer.Bytes(), filename)
			if err != nil {
				log.Printf("Warning: Failed to upload to B2: %v", err)
			}
		}

	case "console":
		fallthrough
	default:
		for _, msg := range messages {
			fmt.Printf("[%s] %s (%s): %s\n",
				msg.Date.Format("2006-01-02 15:04:05"),
				msg.Channel, msg.User, msg.Text)
		}
		// No B2 upload for console output
	}

	return nil
}