## Run testcases

`HttpRunner` can run testcases in diverse ways.

You can run single testset by specifying testset file path.

```text
$ httprunner filepath/testcase.yml
```

You can also run several testsets by specifying multiple testset file paths.

```text
$ httprunner filepath1/testcase1.yml filepath2/testcase2.yml
```

If you want to run testsets of a whole project, you can achieve this goal by specifying the project folder path.

```text
$ httprunner testcases_folder_path
```

When you do continuous integration test or production environment monitoring with `Jenkins`, you may need to send test result notification. For instance, you can send email with mailgun service as below.

```text
$ httprunner filepath/testcase.yml --report-name ${BUILD_NUMBER} \
    --mailgun-smtp-username "qa@debugtalk.com" \
    --mailgun-smtp-password "12345678" \
    --email-sender excited@samples.mailgun.org \
    --email-recepients ${MAIL_RECEPIENTS} \
    --jenkins-job-name ${JOB_NAME} \
    --jenkins-job-url ${JOB_URL} \
    --jenkins-build-number ${BUILD_NUMBER}
```
