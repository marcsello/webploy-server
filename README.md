# Webploy server

This is a simple HTTP server that facilitates atomic deployment of static websites on the local file system.

Webploy can manage multiple deployments of multiple sites and change between the live deployment by updating a symlink to it.

Creating, changing and uploading deployments are all handled through the HTTP API. 

## Deployment lifecycle

First new deployments are created as "open". Files can be uploaded to "open" deployments only, and "open" deployments can not be made live.

After all files uploaded, you can mark the deployment as "finished". After finished no more files can be added to a deployment. Only finished deployments can be made live.

The live deployments are not allowed to be deleted. There must be at least one deployment that is set live.

**There is no supported way to revert a "finished" deployment to be "open" again!**

Unfinished deployments are cleaned up after a certain time if they have no activity (can be disabled in the config). 
Similarly old, finished deployments are deleted when new deployments are being finished, this too can be configured.

## Configuration

### Config file

Webploy is configured via a single YAML file. By default, it tries to load it from `/etc/webploy/webploy.conf` 
but this path can be overridden by setting the `WEBPLOY_CONFIG` env-var.

Example config file structure:

```yaml
listen: # optional if you want to change listen defaults
  bind_addr: ":8000"            # optional, defaults to :8000
  enable_tls: true              # optional, defaults to false
  tls_key: "/etc/tls/key.pem"   # required when enable_tls is true
  tls_cert: "/etc/tls/cert.pem" # required when enable_tls is true
authentication: # required, configure the authentication provider(s) to be used. At least one must be configured.
  basic_auth:   # required if you want to use basic auth authentication, leave it out to disable
    htpasswd_file: "/etc/webploy/.htpasswd" # optional, defaults to "/etc/webploy/.htpasswd"
authorization:  # optional if you want to change authorization defaults
  policy_file: "/etc/webploy/policy.csv" # optional, defaults to "/etc/webploy/policy.csv"
sites: # required, managed sites config
  root: "/var/www" # optional, defaults to "/var/www"
  sites: # required, list of managed sites
    - name: "my_site"               # also the name of the subdirectory bellow "root"
      max_history: 2                # optional, max number of old deployments to keep, the oldest ones will be deleted, default 2
      max_open: 2                   # optional, max number of unfinished deployment at the same time, default 2
      max_concurrent_uploads: 10    # optional, max number of concurrent uploads to the same deployment, set 0 for no limit, default 10
      link_name: "live"             # optional, name of the symlink under "root"/"name", default "live"
      go_live_on_finish: true       # optional, make a deployment live automatically after finishing it, default true
      stale_cleanup_timeout: "30m"  # optional, delete unfinished deployment if there was no activity on them after this time, set 0 to disable. default 30m  
      hooks:                        # optional if you want to define hooks
        pre_create: "/path/to/my/hook/script.sh"  # optional, script to be run before creating a new deployment, no default
        pre_finish: "/path/to/my/hook/script.sh"  # optional, script to be run before finishing a deployment, no default
        post_finish: "/path/to/my/hook/script.sh" # optional, script to be run after finishing a deployment, no default
        pre_live: "/path/to/my/hook/script.sh"    # optional, script to be run before setting a deployment as live, no default
        post_live: "/path/to/my/hook/script.sh"   # optional, script to be run after setting a deployment as live, no default
    - name: "my_other_site" # this is a minimal example, only the name is required
```

### Policy

Another file is required to define roles for users. This is stored at `/etc/webploy/policy.csv` by default, but can be changed in the configuration file as described above. 

This is handled by casbin RBAC setup: <https://casbin.org/docs/rbac/> using the model defined in <authorization/model.conf>.

The fields in the CSV file are `type,sub,obj,act`:

 - **type**: `p` for policy, `g` for group assignment
 - **sub**: subject, the name of the user as returned by the authentication solution
 - **obj**: object: name of the site defined in the config as `name`
 - **act**: action, possible values are defined in <authorization/act_const.go>

Example of such a policy.csv: 
```csv
p,my_site_deployer,my_site,create-deployment
p,my_site_deployer,my_site,upload-self
p,my_site_deployer,my_site,finish-self

g,some_user,my_site_deployer

p,some_other_user,my_site,list-deployments
```

### Authentication

Currently, only basic-auth is supported with a simple htpasswd file.

## API

Webploy currently serves the following api endpoints:

- `GET` `sites/:siteName/live`: Get info of the current live deployment
- `PUT` `sites/:siteName/live`: Update the live deployment
- `GET` `sites/:siteName/deployments`: List available deployments
- `GET` `sites/:siteName/deployments/:deploymentID`: Get deployment information
- `POST` `sites/:siteName/deployments`: Create a new deployment (you can set "meta" here)
- `POST` `sites/:siteName/deployments/:deploymentID/upload`: Upload a single file to a deployment (the request body is the file as-is, file name must be set by the `X-Filename` header.)
- `POST` `sites/:siteName/deployments/:deploymentID/uploadTar`: Upload files in a TAR archive to the deployment (only regualar files will be extracted)
- `POST` `sites/:siteName/deployments/:deploymentID/finish`: Mark a deployment as finished

Refer to <api/api.go> if something seems out of place.

## Hooks

You can add Hooks to specific site or deployment lifecycle events. Info will be provided to the hooks via both arguments and as envvars.

Arguments: The first argument is the id of the lifecycle event, you may find these in <hooks/ids.go>. The second argument may be the deployment path if applicable (not applicable for `pre-create`).

The following envvars may be available, if applicable:
 - `WEBPLOY_HOOK` same as the first argument, the id of the lifecycle event.
 - `WEBPLOY_USER` the user whose action triggered this hook
 - `WEBPLOY_SITE` the name of the site this hook is triggered for
 - `WEBPLOY_SITE_PATH` the path of the site this hook is triggered for (usually `/var/www/my_site`)
 - `WEBPLOY_SITE_CURRENT_LIVE` the id of the current live deployment for the site this hook is triggered for
 - `WEBPLOY_DEPLOYMENT_CREATOR` the name of the user who initially created this deployment
 - `WEBPLOY_DEPLOYMENT_META` the contents of the "meta" provided when the site was created.
 - `WEBPLOY_DEPLOYMENT_PATH` same as the second argument, the path of the deployment this hook is triggered for (not applicable for `pre-create`)
 - `WEBPLOY_DEPLOYMENT_ID` the ID of the current deployment, this hook is triggered for. (not applicable for `pre-create`)

`pre-*` hooks can prevent an action from happening by exiting a non-zero exit code.
If an action is prevented by a hook, Webploy API will return status `424 Failed Dependecy`. `post-*` hooks that return non-zero will do nothing, however.

## Files layout

Under the `root` folder, webploy creates a new folder for each site.

In the folder of each site, Webploy creates folders for each deployment, and a symlink that points to one of these deployments.

In each deployment there is a folder, named `_content` that holds the actual static website content. 
Webploy may put a file in this folder to keep track of some info related to that deployment.

By default, you may configure your static webserver's (Nginx, Apache, lighttpd, Caddy, etc.) site root to the following:

`/var/www/my_site/live/_content`