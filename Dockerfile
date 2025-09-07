FROM mcr.microsoft.com/dotnet/runtime:6.0-alpine

# Install dependencies
RUN apk add --no-cache curl rclone unzip

# Download DiscordChatExporter
WORKDIR /app
RUN curl -Lo DiscordChatExporter.Cli.zip "https://github.com/Tyrrrz/DiscordChatExporter/releases/latest/download/DiscordChatExporter.Cli.linux-x64.zip" \
    && unzip DiscordChatExporter.Cli.zip \
    && rm DiscordChatExporter.Cli.zip \
    && ls -la \
    && chmod +x DiscordChatExporter.Cli

# Create output directory
RUN mkdir -p /output

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]