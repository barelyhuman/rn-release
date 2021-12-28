# rn-release

> react native version sync made simple

## About

The cli simply creates a script that can sync your current version data into your ios `Info.plist` and `build.gradle` for android. You can use the script as is, or you can use the cli to handle the increment process for you..

## Things it can do

- Read and increment version into your package.json (for now, plan on adding similar support for VERSION and .commitlog.release files), this depends on `npm` right now, so it's a dependency for now
- Generate a reusable script for your (will be stored in `.rnrelease/sync-version.sh`), you can execute the script after incrementing the version by yourself in npm

## Install

You can get a binary from the [releases](/releases) page of this repo and add it to a location that's in your unix os's `PATH`

## Usage

```sh
rn-release
```

That's it, no flags right now, though this is for version 0.0.1, might change as we go ahead

## Caveats

- Dependent on npm (v0.0.1)
- Probably more that i'm not aware of yet

## License

[MIT](/LICENSE)
