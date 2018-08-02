# service-broker-ci
GO package that provides a Travis CI framework for testing Service Catalog
Instances.


### Download files
To setup a gate with Travis you need the file ```.travis.yml``` checked into
your repo.
```bash
curl -o .travis.yml https://raw.githubusercontent.com/rthallisey/service-broker-ci/master/travis.yml
```

In order to allow Travis to install go in the gate, there needs to be a go file
in your repo.
```bash
curl -O https://raw.githubusercontent.com/rthallisey/service-broker-ci/master/travis.go
```

Finally, curl an example config.yaml to help guide you.
```
curl -O https://raw.githubusercontent.com/rthallisey/service-broker-ci/master/config.yaml
```


### Creating the ClusterServiceInstance template
At this point, you have already written your apb. Let's create the
ServiceInstance template that will launch your apb.

When using the [apb tool](https://github.com/ansibleplaybookbundle/ansible-playbook-bundle), run
```apb serviceinstace``` to generate a ServiceInstance template. You will need
to edit parameters and select a plan before it's an acceptable resource.


### Naming your resource
It's required that the resource being created shares the name with the
ServiceInstance.  The matching names are used to identify which resource
will receive the bind data.


### Config.yaml Syntax
The file ```config.yaml``` will hold the instructions for running the gate.
Inside config.yaml there are five API KEYS allowed:
- provision     |  Create an application
- bind          |  Connect an application
- unbind        |  Delete an application connection
- deprovision   |  Delete an application
- verify        |  Verify an action succeeded

They are expected to be in the format:
```yaml
<API KEYS>: <FILE>
```

The _FILE_ field accepts a valid git repo ```rthallisey/service-broker-ci/postgresql```
of any APB or a local file/script ```wait-for-resources.sh```. The Four API
Keys: provision , bind, unbind, and deprovision expect a template so the .yaml
extention will be added automatically by the framework.


##### Verify
Verify is used to check if an action is successful.  Use it to determine whether
your APB lifecycle worked as expected.  Verify accepts a script from git repo
```rthallisey/service-broker-ci/wait-for-resource.sh``` or a local script
```wait-for-resource.sh```.

The Verify API Key is also a shell. It can run any shell command and return the
output.
```yaml
verify: oc get pods
```

The Verify action is optional, but it's highly recommended you use it to verify
that your apb worked correctly.
Example verification scripts:
 - https://github.com/fusor/service-broker-ci/blob/master/wait-for-resource.sh
 - https://github.com/fusor/service-broker-ci/blob/master/verify-mediawiki-postgresql.sh


### Config file format
Templates used by provision, bind, unbind, and deprovision are expected to be in
the ```templates``` directory. Everything else uses the full path provided.
```bash
.
|── templates
│   ├── mediawiki123.yaml
│   ├── postgresql-mediawiki123-bind.yaml
│   └── postgresql.yaml
```


##### Using Local Paths
The config file accepts local paths to scripts and templates.

Every template will be searched for in the ```templates``` directory locally
while other scripts will use the top level directory.
```yaml
provision: mediawiki123
verify: wait-for-resource.sh create pod mediawiki
```


##### Matching Paths
When describing a path to a template, that path will be the key used to identify
which app is being acted upon.

For example, to provision and deprovision the same postgresql app, use matching
paths.
```yaml
provision: postgresql
verify: wait-for-resource.sh create pod postgresql

deprovision: postgresql
verify: wait-for-resource.sh delete pod postgresql
```


### Bind Ordering
There are two applications that are used in a bind, the **bindApp** and the
**bindTarget**. The bindApp is the application that will be binded to another
application. If I say: "I want to bind postgresql to mediawiki". Then, the
bindApp will be postgresql. The application that is being binded to, or is
receiving the bind credentials, is the bindTarget. Mediaiwiki is the bindTarget
from the example.

In config.yaml, the bindTarget is determined by the first application
provisioned that's not the same as the bindApp.

This will bind postgresql to mediawiki123.
```yaml
provision: rthallisey/service-broker-ci/mediawiki123
provision: rthallisey/service-broker-ci/postgresql

bind: rthallisey/service-broker-ci/postgresql
```

You can do multiple bindings with different applications where the next bind
call will bind to the next availble provisioned application.

This will bind postgresql to mediawiki123 and mariadb to elasticsearch.
```yaml
provision: rthallisey/service-broker-ci/mediawiki123
provision: rthallisey/service-broker-ci/elasticsearch

provision: rthallisey/service-broker-ci/postgresql
provision: rthallisey/service-broker-ci/mariadb

bind: rthallisey/service-broker-ci/postgresql
bind: rthallisey/service-broker-ci/mariadb
```
