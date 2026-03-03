auth0-create-api:
    auth0 apis create \
    --name {{ project_name }} \
    --identifier "$SAVETOINK_AUTH0_AUDIENCE" \
    --signing-alg "RS256"

auth0-destroy-api:
    auth0 apis delete --force "$SAVETOINK_AUTH0_AUDIENCE"
