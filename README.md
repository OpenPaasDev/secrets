# GPG based secrets management for Git projects
OpenPaaS `secrets` allows you to manage secrets in your Git repository, encrypted by GPG.

You can install it by running `go install github.com/OpenPaasDev/secrets@latest` if you have Go installed and setup, then just run `secrets`.

## Initialising a secrets system
To initialise an environment, and your private key, should you not have one, run:

```
secrets init -b [baseDir for secrets] -e [environment, like dev/prod]
```
This will setup your private key in `~/.openpaas` and a passphrase if you do not have one, as well as create the folder, and add your public key into the environment specific folder.

`secrets` allows you to have different people with privileges to secrets in different ennvironments, based on whose public keys are present in the environment specific `pubkey` folder.

## Adding secrets

```
secrets add -b [baseDir] -e [environment] -n [name of secret]
```

This will prompt you to add the secret. Currently only single line secrets are supported.

## Adding more users with access
To add a new user, simply have them run the initialise workflow, commit their changes, then someone with existing access can run the following command to re-encrypt all secrets, granting access to the new user:

```
secrets refresh -b [baseDir] -e [environment]
```

## Viewing/accessing secrets
Currently only dumping secrets to a file is supported:

```
secrets env -b [baseDir] -e [environment] -o [output file]
```

This command will put all secrets for the environment in the output file in the `export [secretName]="value"` format.
