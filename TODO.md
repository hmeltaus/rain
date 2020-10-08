# To do

* `cat` should do local files too

* add recursive flag to `cat`. Grabs nested templates.

* If `aws` is not installed or doesn't run, use decreased functionality
    * No `package` or `verify`

* Add `aws cloudformation verify-template` to `rain check`

* Use changeset information to be able to display To Do / In Progress / Done during deploy

* Fix bug where using a previous parameter value from a rolled-back stack causes rain to fail deployment

* Handle "allowed values" in parameters

* Fix the bug where failed template validation results in a crash

* Allow watch to watch non-existent or unchanging stacks indefinitely

* Add "event" mode for watch (and related operations). Spit out YAML docs containing changes as they occur.

* If a token expires during an operation (for example watching a stack), rain should refresh it rather than die

* Diff formatting does not handle multi-line strings very well

* When watching a stack, show seconds since beginning of operation, not seconds since beginning of launching the `watch` command - this might be too hard to determine

* `deploy`
    * Add global Include feature - with warning
    * Ensure update count reflects everything that has changed
    * Detect whether a deployment requires capabilities rather than automatically applying them
    * Allow deploying over a stack that's REVIEW_IN_PROGRESS by killing the changeset?
    * Show details from nested stacks while deploying
    * Handle deploying from a template URL

## Other ideas

* Multiple deployments. Use a rain.yaml to specify multiple stacks in multiple regions/accounts.
* `doc` - load documentation for a resource type
* `minify` - try hard to get a template below the size limit
* Do template parameter validation (especially multiple-template stacks - checking clashing outputs etc.)
    * S3 buckets that exist or can't be created (e.g. recent deleted bucket with same name)
    * Certificates that don't exist in the correct region (e.g. non us-east-1)
    * Mismatching or existing "CNAMEs" for CloudFront distros
* Blueprints (higher level constructs - maybe from CDK)
* Store metadata in template?
    * stack name
* Magically add tags?
    * Commit ID
