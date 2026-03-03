setup-mailjet-webhook:
    curl -s \
        -X POST \
        --user "${SAVETOINK_MAILJET_API_KEY}:${SAVETOINK_MAILJET_API_SECRET}" \
        -H 'Content-Type: application/json' \
        -d "{\"Version\": 2, \"EventType\":\"bounce\",\"Status\":\"alive\",\"Url\":\"https://${SAVETOINK_DOMAIN}/v1/webhooks/mailjet?secret=${SAVETOINK_MAILJET_WEBHOOK_SECRET}\"}" \
        https://api.mailjet.com/v3/REST/eventcallbackurl | jq .
