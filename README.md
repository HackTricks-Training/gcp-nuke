# gcp-nuke

Remove all resources from a GCP project.

This project is heavily borrowed from [aws-nuke](https://github.com/rebuy-de/aws-nuke). The goal was to make the interface and operation as consistent with _aws-nuke_ as possible while taking into account the differences between AWS and GCP.

> **Development Status** _gcp-nuke_ is beta and as such it is highly likely that not all GCP resources are covered by it. Be encouraged to add missing resources and create a Pull Request or to create an [Issue](https://github.com/cldcvr/gcp-nuke/issues/new).

## Caution!

Be aware that _gcp-nuke_ is a very destructive tool, hence you have to be very careful while using it. Otherwise you might delete production data.

**We strongly advise you to not run this application on any GCP project, where you cannot afford to lose all resources.**

To reduce the blast radius of accidents, there are some safety precautions:

1. By default _gcp-nuke_ only lists all nukeable resources. You need to add
   `--no-dry-run` to actually delete resources.
2. _gcp-nuke_ asks you twice to confirm the deletion by entering the project ID (name). The first time is directly after the start and the second time after listing all nukeable resources.
3. The config file contains a project-restricted-list field. If the GCP project you want to nuke is part of this blocklist, _gcp-nuke_ will abort. It is recommended, that you add every production project to this blocklist.
4. To ensure you don't just ignore the blocklisting feature, the blocklist must contain at least one project ID.
5. The config file contains project specific settings (e.g. filters). The project you want to nuke must be explicitly listed there.
6. To ensure to not accidentally delete a random account, it is required to specify a config file. It is recommended to have only a single config file and add it to a central repository. This way the project restricted list ia easier to manage and keep up to date.

Feel free to create an issue, if you have any ideas to improve the safety procedures.

## Use Cases

- We are testing our [Terraform](https://www.terraform.io/) code with Jenkins.
  Sometimes a Terraform run fails during development and messes up the project.
  With _gcp-nuke_ we can simply clean up the failed project so it can be reused for the next build.
- Our platform developers have their own GCP projects where they can create their own Kubernetes clusters for testing purposes. With _gcp-nuke_ it is very easy to clean up these projects at the end of the day and keep the costs low.

## Usage

At first you need to create a config file for _gcp-nuke_. This is a minimal one:

```yaml
project-restricted-list:
  - a-prod-env

projects:
  my-test-project:
  locations:
    - global
    - us-east1
```

With this config we can run _gcp-nuke_:

```
$ gcp-nuke -c config/nuke-config.yml my-test-project
gcp-nuke version v1.0.39.gc2f318f - Mon May 8 16:26:42 EDT 2018 - c2f318f37b7d2dec0e646da3d4d05ab5296d5bce

Do you really want to nuke the project with the ID my-test-project?
Do you want to continue? Enter project ID to continue.
> my-test-project
<Need example>
Would delete these resources. Provide --no-dry-run to actually destroy resources.
```

As we see, _gcp-nuke_ only lists all found resources and exits. This is because the `--no-dry-run` flag is missing. Also it wants to delete the
administrator. We don't want to do this, because we use this user to access
our account. Therefore we have to extend the config so it ignores this user:

```yaml
project-restricted-list:
  - a-prod-env

projects:
  my-test-project:
  locations:
    - global
    - us-east1
  filters:
    IAMUser:
      - "my-user"
    IAMUserPolicyAttachment:
      - "my-user -> AdministratorAccess"
    IAMUserAccessKey:
      - "my-user -> ABCDEFGHIJKLMNOPQRST"
```

```
$ aws-nuke -c config/nuke-config.yml --profile aws-nuke-example --no-dry-run
aws-nuke version v1.0.39.gc2f318f - Fri Jul 28 16:26:41 CEST 2017 - c2f318f37b7d2dec0e646da3d4d05ab5296d5bce

Do you really want to nuke the account with the ID 000000000000 and the alias 'aws-nuke-example'?
Do you want to continue? Enter account alias to continue.
> aws-nuke-example

eu-west-1 - EC2DHCPOption - 'dopt-bf2ec3d8' - would remove
eu-west-1 - EC2Instance - 'i-01b489457a60298dd' - would remove
eu-west-1 - EC2KeyPair - 'test' - would remove
eu-west-1 - EC2NetworkACL - 'acl-6482a303' - cannot delete default VPC
eu-west-1 - EC2RouteTable - 'rtb-ffe91e99' - would remove
eu-west-1 - EC2SecurityGroup - 'sg-220e945a' - cannot delete group 'default'
eu-west-1 - EC2SecurityGroup - 'sg-f20f958a' - would remove
eu-west-1 - EC2Subnet - 'subnet-154d844e' - would remove
eu-west-1 - EC2Volume - 'vol-0ddfb15461a00c3e2' - would remove
eu-west-1 - EC2VPC - 'vpc-c6159fa1' - would remove
eu-west-1 - IAMUserAccessKey - 'my-user -> ABCDEFGHIJKLMNOPQRST' - filtered by config
eu-west-1 - IAMUserPolicyAttachment - 'my-user -> AdministratorAccess' - [UserName: "my-user", PolicyArn: "arn:aws:iam::aws:policy/AdministratorAccess", PolicyName: "AdministratorAccess"] - would remove
eu-west-1 - IAMUser - 'my-user' - filtered by config
Scan complete: 13 total, 8 nukeable, 5 filtered.

Do you really want to nuke these resources on the account with the ID 000000000000 and the alias 'aws-nuke-example'?
Do you want to continue? Enter account alias to continue.
> aws-nuke-example

eu-west-1 - EC2DHCPOption - 'dopt-bf2ec3d8' - failed
eu-west-1 - EC2Instance - 'i-01b489457a60298dd' - triggered remove
eu-west-1 - EC2KeyPair - 'test' - triggered remove
eu-west-1 - EC2RouteTable - 'rtb-ffe91e99' - failed
eu-west-1 - EC2SecurityGroup - 'sg-f20f958a' - failed
eu-west-1 - EC2Subnet - 'subnet-154d844e' - failed
eu-west-1 - EC2Volume - 'vol-0ddfb15461a00c3e2' - failed
eu-west-1 - EC2VPC - 'vpc-c6159fa1' - failed
eu-west-1 - S3Object - 's3://rebuy-terraform-state-138758637120/run-terraform.lock' - triggered remove

Removal requested: 2 waiting, 6 failed, 5 skipped, 0 finished

eu-west-1 - EC2DHCPOption - 'dopt-bf2ec3d8' - failed
eu-west-1 - EC2Instance - 'i-01b489457a60298dd' - waiting
eu-west-1 - EC2KeyPair - 'test' - removed
eu-west-1 - EC2RouteTable - 'rtb-ffe91e99' - failed
eu-west-1 - EC2SecurityGroup - 'sg-f20f958a' - failed
eu-west-1 - EC2Subnet - 'subnet-154d844e' - failed
eu-west-1 - EC2Volume - 'vol-0ddfb15461a00c3e2' - failed
eu-west-1 - EC2VPC - 'vpc-c6159fa1' - failed

Removal requested: 1 waiting, 6 failed, 5 skipped, 1 finished

--- truncating long output ---
```

As you see _aws-nuke_ now tries to delete all resources which aren't filtered,
without caring about the dependencies between them. This results in API errors
which can be ignored. These errors are shown at the end of the _aws-nuke_ run,
if they keep to appear.

_aws-nuke_ retries deleting all resources until all specified ones are deleted
or until there are only resources with errors left.

### GCP Credentials

There are two ways to authenticate _gcp-nuke_ - using application default credentials (ADC) or a service account. To use ADC, you just need to authenticate with the gcloud SDK (gcloud auth application-default login). For service account, create the service account and download the key JSON file. Use the --keyfile option to specify the location of that file.

To use _application default credentials_ no command line flags are required. You just need to be authenticated with the SDK prior to running _gcp-nuke_. See [using ADC](https://cloud.google.com/docs/authentication/client-libraries) for more information.

To use a _service account_, the command line flag `--keyfile` is required. Specify the location of the download key JSON file with this flag.

### Specifying Resource Types to Delete

_gcp-nuke_ suppots a subset of resources for deletion. Over time, more resources will be supported and you might want to restrict which resources to process/delete. There are multiple ways to configure this.

The first method is filters, which was mentioned above. This requires you to know the
identifier of each resource or some other metadata about it. It is also possible to prevent whole resource types (eg `Bucket`) from getting deleted with two methods.

- The `--target` flag limits nuking to the specified resource types.
- The `--exclude` flag prevent nuking of the specified resource types.

It is also possible to configure the resource types in the config file like in
these examples:

```yaml
---
project-restricted-list:
  - a-prod-env

resource-types:
 # only nuke resources of these types
  targets:
  - Bucket
  - BucketObject

projects:
  my-test-project:
  locations:
    - global
    - us-east1
```

```yaml
---
project-restricted-list:
  - a-prod-env

resource-types:
 # don't nuke resources of these types
  targets:
  - Bucket
  - BucketObject

projects:
  my-test-project:
  locations:
    - global
    - us-east1
```

If targets are specified in multiple places (eg CLI and account specific), then
a resource type must be specified in all places. In other words each
configuration limits the previous ones.

If an exclude is used, then all its resource types will not be deleted.

**Hint:** You can see all available resource types with this command:

```
$ gcp-nuke resource-types
```

### Filtering Resources

It is possible to filter this is important for not deleting the current user
for example or for resources like S3 Buckets which have a globally shared
namespace and might be hard to recreate. Currently the filtering is based on
the resource identifier. The identifier will be printed as the first step of
_aws-nuke_ (eg `i-01b489457a60298dd` for an EC2 instance).

**Note: Even with filters you should not run aws-nuke on any AWS account, where
you cannot afford to lose all resources. It is easy to make mistakes in the
filter configuration. Also, since aws-nuke is in continous development, there
is always a possibility to introduce new bugs, no matter how careful we review
new code.**

The filters are part of the account-specific configuration and are grouped by
resource types. This is an example of a config that deletes all resources but
the `admin` user with its access permissions and two access keys:

```yaml
---
regions:
  - global
  - eu-west-1

account-blocklist:
  - 1234567890

accounts:
  0987654321:
    filters:
      IAMUser:
        - "admin"
      IAMUserPolicyAttachment:
        - "admin -> AdministratorAccess"
      IAMUserAccessKey:
        - "admin -> AKSDAFRETERSDF"
        - "admin -> AFGDSGRTEWSFEY"
```

Any resource whose resource identifier exactly matches any of the filters in
the list will be skipped. These will be marked as "filtered by config" on the
_aws-nuke_ run.

#### Filter Properties

Some resources support filtering via properties. When a resource support these
properties, they will be listed in the output like in this example:

```
global - IAMUserPolicyAttachment - 'admin -> AdministratorAccess' - [RoleName: "admin", PolicyArn: "arn:aws:iam::aws:policy/AdministratorAccess", PolicyName: "AdministratorAccess"] - would remove
```

To use properties, it is required to specify a object with `properties` and
`value` instead of the plain string.

These types can be used to simplify the configuration. For example, it is
possible to protect all access keys of a single user:

```yaml
IAMUserAccessKey:
  - property: UserName
    value: "admin"
```

#### Filter Types

There are also additional comparision types than an exact match:

- `exact` – The identifier must exactly match the given string. This is the default.
- `contains` – The identifier must contain the given string.
- `glob` – The identifier must match against the given [glob
  pattern](<https://en.wikipedia.org/wiki/Glob_(programming)>). This means the
  string might contains wildcards like `*` and `?`. Note that globbing is
  designed for file paths, so the wildcards do not match the directory
  separator (`/`). Details about the glob pattern can be found in the [library
  documentation](https://godoc.org/github.com/mb0/glob).
- `regex` – The identifier must match against the given regular expression.
  Details about the syntax can be found in the [library
  documentation](https://golang.org/pkg/regexp/syntax/).
- `dateOlderThan` - The identifier is parsed as a timestamp. After the offset is added to it (specified in the `value` field), the resulting timestamp must be AFTER the current
  time. Details on offset syntax can be found in
  the [library documentation](https://golang.org/pkg/time/#ParseDuration). Supported
  date formats are epoch time, `2006-01-02`, `2006/01/02`, `2006-01-02T15:04:05Z`,
  `2006-01-02T15:04:05.999999999Z07:00`, and `2006-01-02T15:04:05Z07:00`.

To use a non-default comparision type, it is required to specify an object with
`type` and `value` instead of the plain string.

These types can be used to simplify the configuration. For example, it is
possible to protect all access keys of a single user by using `glob`:

```yaml
IAMUserAccessKey:
  - type: glob
    value: "admin -> *"
```

#### Using Them Together

It is also possible to use Filter Properties and Filter Types together. For
example to protect all Hosted Zone of a specific TLD:

```yaml
Route53HostedZone:
  - property: Name
    type: glob
    value: "*.rebuy.cloud."
```

#### Inverting Filter Results

Any filter result can be inverted by using `invert: true`, for example:

```yaml
CloudFormationStack:
  - property: Name
    value: "foo"
    invert: true
```

In this case _any_ CloudFormationStack **_but_** the ones called "foo" will be
filtered. Be aware that _aws-nuke_ internally takes every resource and applies
every filter on it. If a filter matches, it marks the node as filtered.

#### Filter Presets

It might be the case that some filters are the same across multiple accounts.
This especially could happen, if provisioning tools like Terraform are used or
if IAM resources follow the same pattern.

For this case _aws-nuke_ supports presets of filters, that can applied on
multiple accounts. A configuration could look like this:

```yaml
---
regions:
  - "global"
  - "eu-west-1"

account-blocklist:
  - 1234567890

accounts:
  555421337:
    presets:
      - "common"
  555133742:
    presets:
      - "common"
      - "terraform"
  555134237:
    presets:
      - "common"
      - "terraform"
    filters:
      EC2KeyPair:
        - "notebook"

presets:
  terraform:
    filters:
      S3Bucket:
        - type: glob
          value: "my-statebucket-*"
      DynamoDBTable:
        - "terraform-lock"
  common:
    filters:
      IAMRole:
        - "OrganizationAccountAccessRole"
```

## Install

### Use Released Binaries

The easiest way of installing it, is to download the latest
[release](https://github.com/cldcvr/gcp-nuke/releases) from GitHub.

#### Example for Linux Intel/AMD

Download and extract
TODO - fix this link
`$ wget -c https://github.com/cldcvr/gcp-nuke/releases/download/v2.16.0/aws-nuke-v2.16.0-linux-amd64.tar.gz -O - | sudo tar -xz -C $HOME/bin`

TODO fix this
Run
`$ aws-nuke-v2.16.0-linux-amd64`

### Compile from Source

To compile _gcp-nuke_ from source you need a working
[Golang](https://golang.org/doc/install) development environment. The sources
must be cloned to `$GOPATH/src/github.com/cldcvr/gcp-nuke`.

Also you need to install [staticcheck](https://github.com/dominikh/go-tools#installation) and [GNU Make](https://www.gnu.org/software/make/).

Then you just need to run `make build` to compile a binary into the project directory or `make install` go install _gcp-nuke_ into `$GOPATH/bin`.

### Docker

You can run _gcp-nuke_ with Docker by using a command like this:

```bash
$ docker run \
    --rm -it \
    -v /full-path/to/nuke-config.yml:/home/gcp-nuke/config.yml \
    -v $HOME/.config/gcloud:/home/gcp-nuke/.config/gcloud \
    cldcvr/gcp-nuke:latest \
    --config /home/gcp-nuke/config.yml \
    --project gcp-project-name
```

To make it work, you need to adjust the paths for the GCP config and the
_gcp-nuke_ config.

The above example uses the application-default credential method. To use a service account keyfile:

```bash
$ docker run \
    --rm -it \
    -v /full-path/to/nuke-config.yml:/home/gcp-nuke/config.yml \
    -v /full-path/to/service-acct-json.json:/home/gcp-nuke/sa.json \
    cldcvr/gcp-nuke:latest \
    --config /home/gcp-nuke/config.yml \
    --project gcp-project-name \
    --keyfile /home/gcp-nuke/sa.json
```

## Testing

### Unit Tests

To unit test _gcp-nuke_, some tests require [gomock](https://github.com/golang/mock) to run.
This will run via `go generate ./...`, but is automatically run via `make test`.
To run the unit tests:

```bash
make test
```

## Contribute

You can contribute to _gcp-nuke_ by forking this repository, making your changes and creating a Pull Request against our repository. If you are unsure how to solve a problem or have other questions about a contributions, please create a GitHub issue.
