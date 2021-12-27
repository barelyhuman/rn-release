# Notes

- Create a `.rnrelease` folder
- Add in save variables data for the `podfile` and `gradlefile` location
- Use the compiled template to execute the version sync
- Create a new compiled template if the above isn't found
- use `semver` to get the values
- use `npm version` to increment it in the `package.json`
- use bash script for making the change in the actual files
