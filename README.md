# gchl

A Go-written Changelog Generator. Create Changelogs based on GitHub pull requests.

## Installation

```
go get -v
go install
```

## Usage

You will need a github personal access token for api calls, create one [here](https://github.com/settings/tokens).
You can pass the token to `gchl` as a flag `--token` or via env variable `GCHL_GITHUB_TOKEN`

Navigate into a repository and get a Changelog of all changes in form of merged PRs between two branches / tags

```
gchl between v1.2 v1.0
```

or since between `HEAD` and another reference

```
gchl since v1.0
```

### Get release notes via PR message annotation

In your pull request use a markdown code block annotated with `release-note` (Don't copy paste the example below as it uses `'` ;))

```
'''release-notes
This text will be visible in changelog
'''
```

and run

```
gchl between v1.2 v1.0 --release-notes
```

The block can also contain a type of a change, which will be used to group the changes in the log, e.g.:

```
'''release-notes bugfix
The important functionality has been fixed
'''
```

will result in the message being grouped under `bugfix:`. Changes without a type specified will be grouped under `misc:`.

## Overview

```
NAME:                                               
   gchl - A Go-written Changelog Generator - Generate Changelogs, based on GitHub PRs                    

USAGE:                                              
   gchl [global options] command [command options] [arguments...]                                        

VERSION:                                            
   v0.1                                             

AUTHOR:                                             
   Christian Bargmann <chris@cbrgm.de>              

COMMANDS:                                           
     between  Create a changelog for changes between to references.                                      
     since    Create a changelog for changes since reference.                                            
     help, h  Shows a list of commands or help for one command                                           

GLOBAL OPTIONS:                                     
   --for-version value, -f value     Specify a version name that will be shown in changelog output (default: "v0.0.0")                                                                                             
   --repository value, --repo value  The file path to the directory containing the git repository to be used (default: "/home/chris/go/src/github.com/kubermatic/kubermatic")                                      
   --remote value, -r value          The remote github repository url                                    
   --token value, -t value           Your personal access token provided by GitHub Inc. See: https://github.com/settings/tokens [$GCHL_GITHUB_TOKEN]                                                               
   --help, -h                        show help      
   --version, -v                     print the version
```
