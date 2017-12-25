## Yet Another Github Pivotal Tracker Sync

### The Rules

1. Creating/updating of Github (GH) issues will create/update a Pivotal Tracker (PT) story with the same title & issue body
1. Creating/updating of Pivotal Tracker stories will create/update a Github issue with the same name & description
1. GH issue body will prefix with a hyperlink to PT story
1. PT story description will prefix with a hyperlink to GH issue
1. Closing a GH issue (e.g. by merging a related pull request) will `Finish` the associated PT story; if story was not estimated, it will be accepted as a chore
1. Rejecting a PT story will re-open the associated GH issue
1. Accepting a PT story will close the associated GH issue
1. Deleting a PT story will disassociate the GH issue; appending of `[no story]` suffix to issue title prevents it from syncing to PT

Non-Goals: Comments are not and will not be synchronised. Do not discuss on Pivotal Tracker.

### Usage

```
make
```

will compile a `build/server` binary that you can run.

#### Environment variables

1. `PORT` defines the port that the http server will listen on
2. `SECRET` is a UUID string, e.g. `c1626442-0327-40a6-a830-c5517d6782d2`

#### Getting started

1. Start the server and visit the URL, e.g.

    ```
    PORT=3000 SECRET=c1626442-0327-40a6-a830-c5517d6782d2 ./build/server
    ````
    > then go to http://localhost:3000/

2. You'll see 2 very ugly html forms
3. Fill in the details and click submit to obtain the webhook urls for your GH repo and PT project

    > NOTE: your server must be on a network accessible *from* github.com and pivotaltracker.com; http://localhost:3000/ won't work

4. One deployment can support multiple GH repo and PT projects, since the details are embedded in the webhook urls (instead of configured centrally on the server)
