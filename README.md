**************************************************************************
# hugger
Huggingface colourful client in terminal.

## Build
In order to build Hugger for the current platform you can run:
```
make build
```
If you want to build Hugger for any other platform you must specify os, for example:
```
make build-linux
```
or
```
make build-windows
```

The program in located in /cmd-cli directory.

## Install
In UNIX-like systems you must move cmd-cli/hugger to any directory in `$PATH` or you can optionally include hugger in your `$PATH` environmental variable:
```
$ export PATH=$PATH:`pwd`/cmd-cli/hugger
```

In Windows you must do almost the same: just include binary in your `%PATH%`.

## Usage
Examples of usage:
```
# in cmd-cli/ directory
# show help menu (and fancy banner)
$ ./hugger -h

# download files from repo
$ ./hugger download -repo-id username/dataset-example -filenames my_dataset_0001.parquet -repo-type dataset -token hf_<your_token_here>


# upload files from to repo
$ ./hugger upload -repo-id username/dataset-example -filenames my_dataset_0001.parquet,my_dataset_0002.parquet -repo-type dataset -token hf_<your_token_here>

# perform actions on files in repo:
$ ./hugger repo-files -repo-id <your_repo_id> -action delete -file unused_file.test -token hf_<your_token_here>

# show meta info about repository
$ ./hugger meta -repo-id <your_repo_id> -repo-type model -token hf_<your_token_here>
```

## Final words
I hope you like this application and it will be useful to you. If you know how to wider its functionality or make code better,
feel free to open an issue or a pull request.

**************************************************************************
