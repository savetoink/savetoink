# Test deployed Lambda function with article URL
smoketest-create-article URL:
    curl --silent \
      -X POST http://localhost:8080/v1/articles \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $SAVETOINK_API_KEY" \
      -d "{\"url\": \"{{ URL }}\", \"tags\":\"test\"}" | jq .

smoketest-send-article ID:
    curl --silent \
      -X POST http://localhost:8080/v1/articles/{{ ID }}/send \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $SAVETOINK_API_KEY" | jq .

smoketest-get-article ID:
    curl --silent \
      -X GET http://localhost:8080/v1/articles/{{ ID }} \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $SAVETOINK_API_KEY" | jq .

smoketest-get-articles PAGE="1":
    curl --silent \
      -X GET http://localhost:8080/v1/articles?page={{ PAGE }} \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $SAVETOINK_API_KEY" | jq .

smoketest-delete-article ID:
    curl --silent \
      -X DELETE http://localhost:8080/v1/articles/{{ ID }} \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $SAVETOINK_API_KEY" | jq .

smoketest-auth-token-exchange CODE:
    curl --silent \
        -X POST \
        --data '{"code":"{{ CODE }}", "redirect_uri": "http://localhost:5173"}' \
        http://localhost:8080/v1/auth/token | jq .

smoketest-auth-browser-login:
    open "https://${SAVETOINK_AUTH0_DOMAIN}/authorize?response_type=code&client_id=${SAVETOINK_AUTH0_CLIENT_ID}&redirect_uri=http://localhost:5173&scope=openid%20profile%20email&state=test123&audience=${SAVETOINK_AUTH0_AUDIENCE}"

smoketest-get-sends:
    curl --silent \
      -X GET http://localhost:8080/v1/sends \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $SAVETOINK_API_KEY" | jq .
