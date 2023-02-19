# aws-parameter-bulk

Utility to read parameters from AWS Systems Manager (SSM) Parameter Store in bulk and output them in environment-file or json format.
It can read all parameters for a given path, or read a list of single parameters. If the parameters contain json, this can be parsed as single values via a flag. 
It uses your current aws profile to access AWS SSM,
you can supply a different profile if you need to read from a different account.
Have your AWS CLI set up correctly. See below for instructions.


The output can be used as .env in your development workspace, as --from-env in docker, or as Kubernetes secret.

## Install via Homebrew

```bash
$ brew tap gork74/gork74

$ brew install aws-parameter-bulk
```

## Usage

`get` reads names from single values, or from a path recursively.
Use --help for usage and parameters.

````bash
$ aws-parameter-bulk --help

$ aws-parameter-bulk get --help
````


Assuming you have the following structure in SSM,
and the parameters are filled with "valueOfParam1" etc.:  

````
/dev/test/param1
/dev/test/param2
/dev/test/param3
/dev/other/other1
/dev/other/other2
/dev/testextend/param1
/dev/path/param1
/dev/path/subpath/subparam1
someparam1
someparam2
jsonparam1
jsonparam2
````

## Get Path

These are the outputs you can create for a path variable.
Note that the last part of the path will be printed in upper case, if you supply the --upper flag.
To be a valid ENV Identifier the output has to use this format: `[a-zA-Z_][a-zA-Z0-9_]*`


````bash
$ aws-parameter-bulk get /dev/test --upper
PARAM1=valueOfParam1
PARAM2=valueOfParam2
PARAM3=valueOfParam3
````

## Get Path without recursion

Paths will be read recursively by default, to turn that off supply the --norecursive flag.

````bash
$ aws-parameter-bulk get /dev/path --upper --norecursive
PARAM1=valueOfParam1
````
Without the flag you would get:
````bash
$ aws-parameter-bulk get /dev/path --upper
PARAM1=valueOfParam1
SUBPARAM1=valueOfSubParam1
````

## Get Path with path prefix

If you want to add the full path to the output, use the --prefixpath flag.

````bash
$ aws-parameter-bulk get /dev/path --prefixpath
/dev/path/param1=valueOfParam1
/dev/path/subpath/subparam1=valueOfSubParam1
````

## Get Path with normalized path prefix

If you want to add the full path to the output, with underscores as separator (to make it bash variable compliant),
use the --prefixnormalizedpath flag. The first slash of the path is not replaced with an underscore, it will be removed.

````bash
$ aws-parameter-bulk get /dev/path --prefixnormalizedpath
dev_path_param1=valueOfParam1
dev_path_subpath_subparam1=valueOfSubParam1
````

## Get Multiple Paths

You can supply multiple paths:

````bash
$ aws-parameter-bulk get /dev/test,/dev/other --upper
PARAM1=valueOfParam1
PARAM2=valueOfParam2
PARAM3=valueOfParam3
OTHER1=valueOfOther1
OTHER2=valueOfOther2
````

## Overwrite Values

An env file key must be unique, therefore it will be filtered so each key only occurs once.
The last key to appear will be printed out, so this will overwrite /dev/test/param1 with /dev/testextend/param1.
This can be used to first read some default values and overwrite some of them.

````bash
$ aws-parameter-bulk get /dev/test,/dev/testextend --upper
PARAM1=valueOfParamFromExtend1
PARAM2=valueOfParam2
PARAM3=valueOfParam3
````

## JSON Output

Output path parameters as JSON file:

````bash
$ aws-parameter-bulk get /dev/test,/dev/other --upper --outjson
````
````json
{
    "PARAM1": "valueOfParam1",
    "PARAM2": "valueOfParam2",
    "PARAM3": "valueOfParam3",
    "OTHER1": "valueOfOther1",
    "OTHER2": "valueOfOther2"
}
````

## Get Single Parameters

Reading single (non-path) SSM Parameters.

````bash
$ aws-parameter-bulk get someparam1,someparam2 --upper
SOMEPARAM1=valueOfSomeParam1
SOMEPARAM2=valueOfSomeParam2
````

## Get Single Parameters on a path

Reading single SSM Parameters on a path.

````bash
$ aws-parameter-bulk get /dev/test/param1,/dev/test/param2 --upper
PARAM1=valueOfParam1
PARAM2=valueOfParam2
````

## Get Parameters Containing JSON

Reading SSM Parameters containing JSON, parsing and converting them. This also works for path parameters. Each parameter has to be json. 

Assuming this is jsonparam1:
````json
{
  "Json1a": "value1a",
  "Json1b": "value1b"
}
````
And jsonparam2:
````json
{
  "JSON2a": "value2a",
  "JSON2b": "value2b"
}
````

This will be the output:
````bash
$ aws-parameter-bulk get jsonparam1,jsonparam2 --injson --upper
JSON1A=value1a
JSON1B=value1b
JSON2A=value2a
JSON2B=value2b
````

## Saving From .env File To SSM Names

Takes a file in `KEY=value` form, and store each line as name and valie in ssm.

````bash
$ aws-parameter-bulk save .env
NAME1
NAME2
````


## Saving From .env File To SSM Paths

Takes a file in `KEY=value` form, prefixes each key with the given path, and stores it in ssm. 

````bash
$ aws-parameter-bulk save .env /dev/something
/dev/something/PARAM1
/dev/something/PARAM2
````

## Saving From JSON File To SSM Paths

Using a json file as input and storing it to a path

````bash
$ aws-parameter-bulk save .env /dev/something --injson

/dev/something/key1=val1
2021-12-07T22:38:19Z INF pkg/util/awsssm.go:174 > Output: {
  Version: 1
}
/dev/something/key2=val2
2021-12-07T22:38:20Z INF pkg/util/awsssm.go:174 > Output: {
  Version: 1
}
````

## Debugging

Add SSM_LOG_LEVEL=debug

````bash
$ SSM_LOG_LEVEL=debug aws-parameter-bulk get jsonparam1, jsonparam2 --injson --upper
````

# Web UI

Start with parameter "web" to start a web ui on [http://localhost:8888](http://localhost:8888).
Change the listen ip and port with the `--address` flag.

````bash
$ aws-parameter-bulk web

$ aws-parameter-bulk web --address :1234
````


# AWS Setup

https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html

It is important that you set your region in your aws profile.

````bash
$ aws configure
AWS Access Key ID [None]: AKIAIOSFODNN7EXAMPLE
AWS Secret Access Key [None]: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
Default region name [None]: eu-central-1
Default output format [None]: json
````

If you have multiple profiles like this (.aws/config):

````
[default]
account = 11111111111
region = eu-central-1
output = json

[profile other]
account = 2222222222
region = eu-central-1
output = json
source_profile = default
````

You can read the SSM Parameters from the other account like this:

````bash
$ AWS_PROFILE=other aws-parameter-bulk get /dev/test
````
