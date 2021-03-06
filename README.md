# action-update-brewformula

## This is not endorsed by or associated with GitHub, Dependabot, etc.

This action checks for available dependency updates to a repository full of simple [homebrew formulae](https://github.com/Homebrew/homebrew-core/tree/59bffb2cbc55deed9cab44d749da9218d32535f1/Formula).

This is an abandoned tech demo. Compared to https://github.com/thepwagner/action-update-go and https://github.com/thepwagner/action-update-docker the implementation is brittle: based on regular expressions without evaluating the formula as code.
Functionality is similar to https://github.com/thepwagner/action-update-dockerurl: find new versions, find new artifact SHASUMs.

The only novel feature is [optional GPG signature verification](https://github.com/thepwagner/action-update-brewformula/pull/7#issuecomment-783333325) of artifacts: this avoid running a potentially malicious release through the CI process.
