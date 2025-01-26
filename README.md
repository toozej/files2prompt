# files2prompt
LLM Prompt Generator From Local Files

- set up new repository in quay.io web console
    - (DockerHub and GitHub Container Registry do this automatically on first push/publish)
    - name must match Git repo name
    - grant robot user with username stored in QUAY_USERNAME "write" permissions (your quay.io account should already have admin permissions)
- set built packages visibility in GitHub packages to public
    - navigate to https://github.com/users/$USERNAME/packages/container/$REPO/settings
    - scroll down to "Danger Zone"
    - change visibility to public

## changes required to update golang version
- run `./scripts/update_golang_version.sh $NEW_VERSION_GOES_HERE`
