package towerdiscord

import (
	"encoding/json"
)

const baseDiscordURL = "https://discord.com/api/v10"

type HTTPErrorResponse struct {
	Code    int             `json:"code"`
	Errors  json.RawMessage `json:"errors"`
	Message string          `json:"message"`
}

type Auth struct {
	Type  string
	Token string
}

/*
Example Message Embed.

{
  "channel_id": `${context.params.event.channel_id}`,
  "content": `<@here> An error has occurred on service \`Service\` on Scope \`scope\``,
  "tts": false,
  "embeds": [
    {
      "type": "rich",
      "title": `sql driver: failed to open connection to 192.168.1.1`,
      "description": `**Summary**\n```\nfailed on executing GetMessage\n    database down\n    sql driver: failed to open connection to 192.168.1.1\n\nfoo:\n    bar: 2000\n```\n**Context**:\n```\nfoo:\n    {\n        \"bar\": 2000\n    }\n```\n\n**Stack Trace**\n```\n\"failed on executing GetMessage\" on /foo/bar.go:20\n\n\"database down\" on /repo/postgres/get_message.go:45\n\nsql driver: failed to open connection to 192.168.1.1\n````,
      "color": 0xff0000,
      "fields": [
        {
          "name": `Trace ID`,
          "value": `123451234512345123451234512345`,
          "inline": true
        },
        {
          "name": `Transaction ID`,
          "value": `098712340987612345098712340987612345`,
          "inline": true
        },
        {
          "name": `Environment`,
          "value": `development`,
          "inline": true
        },
        {
          "name": `Timestamp`,
          "value": `Thu, 03 Nov 2022 14:47:23 +0000`,
          "inline": true
        },
        {
          "name": `Cooldown`,
          "value": "\u200B",
          "inline": true
        }
      ],
      "timestamp": `2022-11-03T14:47:23.000Z`,
      "author": {
        "name": `Service - scope`,
        "url": `https://example.com`,
        "icon_url": `https://previews.123rf.com/images/valentint/valentint1704/valentint170400245/75400909-examples-icon-examples-website-button-on-white-background-.jpg`
      },
      "footer": {
        "text": `Message Key: `/foo/bar.go:20``
      },
      "url": `https://example.com`
    }
  ]
}
*/
