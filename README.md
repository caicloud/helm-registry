# Helm Registry

Helm Registry stores helm charts in a hierarchy storage structure and provides a function to orchestrate charts form existed charts. The structure is:
```
|- space
  |- chart
    |- version
```
Every space is independent with others. It means the registry can stores same charts (same name with same version) in two spaces.

### Build
Build with `make`:
```
$ cd $ROOT_OF_PROJECT
$ make registry
```
Then you can get binary at `./bin/registry`


### Configuration
Before you start `./bin/registry`, you need to wirte a config file which named `config.yaml` in `./bin`.
Config file is explained below:
```yaml
# The port which the server listen to. Change to any port you like.
listen: ":8099"
# A manager is a charts manager. Now we only support `simple` manager.
manager:
  # The name of charts manager.
  name: "simple"
  # The config of current manager.
  parameters:
    # A manager manages all operations of charts. So it is responsible for sync read and write operationgs.
    # The option indicates which locker the manager will use. Currently we provide a `memory` locker.
    resourcelocker: memory
    # A manager can use many storage backends.
    storagedriver: filesystem
    # The option is a parameter of storage driver `filesystem`. See below `Storage Backends`
    rootdirectory: ./data
```

### Storage Backends
We simply use docker backends as manager storage backends. But now we only have build-in support of `filesystem`.
For more infomation of backends, please refer to [Docker Backends](https://docs.docker.com/registry/storage-drivers/)


### Usage
After registry running, you can manage the registry by a registy client (in `pkg/rest/v1`) or simply use http APIs.
In `pkg/api/v1/descriptor`, you can find all descriptors of these APIs.

### Orchestration
The registry can orchestrate charts by a json config like:
```
{
    "save":{
        "chart":"chart name",           // new chart name
        "version":"1.0.0",              // new chart version
        "description":"description"     // new chart description
    },
    "configs":{                         // configs is the orchestration configuration of new chart
        "package":{                     // package indicates a original chart which new chart is from
            "independent":true,         // if the original chart is a independent chart, the option is true
            "space":"space name",       // space/chart/version indicate where original chart is stored
            "chart":"chart name",
            "version":"version number"
        },
        "_config": {
        // root chart config, these configs will store in values.yaml of new chart.
        },
        "chartB": {                     // rename original chart as `chartB`
            "package":{
                "independent":true,     // for explaining, we call this original chart as `XChart`
                "space":"space name",
                "chart":"chart name",
                "version":"version number"
            },
            "_config": {
                // chartB config
            },
            "chartD":{
                "package":{
                    "independent":false,  // if independent is false, it means the original chart is a subchart of `XChart`
                    "space":"space name",
                    "chart":"chart name",
                    "version":"version number"
                },
                "_config": {
                    // chartD config
                }
            }
        },
        "chartC": {
            "package":{
                "independent":false,
                "space":"space name",
                "chart":"chart name",
                "version":"version number"
            },
            "_config": {
                // chartC config
            }
        }
    }
}

```
