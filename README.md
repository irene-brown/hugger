![Banner](https://github.com/irene-brown/hugger/blob/main/hugger_banner.png?raw=true)
**************************************************************************
# hugger

![License](https://img.shields.io/badge/license-GPLv3-blue.svg?style=for-the-badge)
![Version](https://img.shields.io/badge/version-0.3.0-brightgreen.svg?style=for-the-badge)
![Issues](https://img.shields.io/github/issues/irene-brown/hugger.svg?style=for-the-badge)
![Coverage](https://img.shields.io/codecov/c/github/irene-brown/hugger.svg?style=for-the-badge)


**************************************************************************
Unofficial Hugging Face client with a dash of personality. Hugger brings a touch of colour and flair to your Hugging Face experience, making it easier and more enjoyable to interact with the popular transformer models, datasets, spaces. It also allows to do the same with your private repositories.

## Features
- Colorful interface
- Fast and reliable
- Easy repository interaction and management
- Advanced features like metadata and statistics viewers

## Build && install

> [!NOTE]
> You can download the latest version of application ![here](https://github.com/irene-brown/hugger/releases)

In order to build Hugger for the current platform you can run:
```bash
make build
```
If you want to build Hugger for any other platform you must specify os, for example:
```bash
make build-linux
```
or
```bash
make build-windows
```

The program is located in build directory.

In UNIX-like systems you must move build/hugger to any directory in `$PATH` or you can optionally include Hugger in your `$PATH` environmental variable:
```bash
$ export PATH=$PATH:`pwd`/build/hugger_linux
```

In Windows you must do almost the same: just include binary in your `%PATH%`.

## Usage
Examples of usage:
```bash
# in build/ directory
# show help menu (and fancy banner)
$ ./hugger -h

# download files from repo
$ ./hugger download -repo-id 'username/dataset-example' -filenames my_dataset_0001.parquet -repo-type dataset -token "hf_<your_token_here>"


# upload files from to repo
$ ./hugger upload -repo-id 'username/dataset-example' -filenames my_dataset_0001.parquet,my_dataset_0002.parquet -repo-type dataset -token "hf_<your_token_here>"

# perform actions on files in repo:
# delete file unused_file.test
$ ./hugger repo-files -repo-id '<your_repo_id>' -action delete -file unused_file.test -token "hf_<your_token_here>"
# list files in the / folder of repository
$ ./hugger repo-files -repo-id '<your_repo_id>' -action list -token "hf_<your_token_here>"
# list files in the /model folder of repository
$ ./hugger repo-files -repo-id '<your_repo_id>' -action list -file model -token "hf_<your_token_here>"

# show meta info about repository
$ ./hugger meta -repo-id '<your_repo_id>' -repo-type model -token "hf_<your_token_here>"

# show statistics for dataset
$ ./hugger statistics -repo-id '<your_repo_id>' -token "hf_<your_token_here>"
```

## Contribution

If you'd like to contribute to Hugger, please follow these steps:

- Fork the repository and create a new branch for your feature or bug fix
- Make your changes and commit them with a clear and descriptive commit message
- Open a pull request and describe the changes you've made

The contribution in the following areas are welcome:
- Bug fixes and stability improvements
- New features and model integrations
- UI/UX enhancements and customizations
- Documentation and testing improvements

Thanks for considering a contribution to Hugger!
**************************************************************************
