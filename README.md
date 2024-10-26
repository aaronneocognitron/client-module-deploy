# Asterizm Protocol client off-chain module script

## About

Shell scripts (for [64-bit](./bin/linux_x64) and [32-bit](./bin/linux_x32) systems) developed to simplify the integration of the Asterizm Protocol client off-chain module.

Please check out the [documentation](https://docs.asterizm.io/guides/getting-started/2.-implement-off-chain-module/simple-implementation-shell-script) if you want to see a detailed example of usage.

This script allows to deploy the module only on **Linux**-based servers with `apt` package manager installed (e.g. **Ubuntu**, **Debian**). If you want to deploy the module on other systems (e.g. Windows), you need to manually configure it (see [Default implementation](https://docs.asterizm.io/guides/getting-started/2.-implement-off-chain-module/default-implementation-manual)).

## How to use

First you need to create a configuration file. You can use this template: [config.full.yml](./config.full.yml).

Then based on your system's architecture you need to download the appropriate lunix_x**XX** script: [linux_x64](./bin/linux_x64) or [linux_x32](./bin/linux_x32).

After downloading, you need to run it with **sudo** privileges (required for installing Docker), passing the full path to your configuration file as a parameter:

```bash
sudo ./lunix_xXX -f /path/to/config.yml
```

In case of deploying a test environment (testnet), you need to add an additional flag: `-test`

```bash
sudo ./lunix_xXX -f /path/to/config.yml -test
```

After the script executes successfully, your environment will be configured, and the client's off-chain module will be up and running.
