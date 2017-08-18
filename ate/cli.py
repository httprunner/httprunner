import argparse
import codecs
import logging
import os
import sys
from collections import OrderedDict
import PyUnitReport

from ate import __version__
from ate.task import create_task


def main_ate():
    """ API test: parse command line options and run commands.
    """
    parser = argparse.ArgumentParser(
        description='Api Test Engine.')
    parser.add_argument(
        '-V', '--version', dest='version', action='store_true',
        help="show version")
    parser.add_argument(
        'testset_paths', nargs='*',
        help="testset file path")
    parser.add_argument(
        '--log-level', default='INFO',
        help="Specify logging level, default is INFO.")
    parser.add_argument(
        '--report-name',
        help="Specify report name, default is generated time.")
    parser.add_argument(
        '--failfast', action='store_true', default=False,
        help="Stop the test run on the first error or failure.")

    try:
        from jenkins_mail_py import MailgunHelper
        mailer = MailgunHelper(parser)
    except ImportError:
        mailer = None

    args = parser.parse_args()

    if args.version:
        print("ApiTestEngine version: {}".format(__version__))
        exit(0)

    log_level = getattr(logging, args.log_level.upper())
    logging.basicConfig(level=log_level)

    report_name = args.report_name
    if report_name and len(args.testset_paths) > 1:
        report_name = None
        logging.warning("More than one testset paths specified, \
                        report name is ignored, use generated time instead.")

    results = {}
    subject = "SUCCESS"

    for testset_path in set(args.testset_paths):

        testset_path = testset_path.strip('/')
        task_suite = create_task(testset_path)

        output_folder_name = os.path.basename(os.path.splitext(testset_path)[0])
        kwargs = {
            "output": output_folder_name,
            "report_name": report_name,
            "failfast": args.failfast
        }
        result = PyUnitReport.HTMLTestRunner(**kwargs).run(task_suite)
        results[testset_path] = OrderedDict({
            "total": result.testsRun,
            "successes": len(result.successes),
            "failures": len(result.failures),
            "errors": len(result.errors),
            "skipped": len(result.skipped)
        })

        if len(result.successes) != result.testsRun:
            subject = "FAILED"

    flag_code = 0 if subject == "SUCCESS" else 1
    if mailer and mailer.config_ready:
        mailer.send_mail(subject, results, flag_code)

    return flag_code

def main_locust():
    """ Performance test with locust: parse command line options and run commands.
    """
    try:
        from locust.main import main
    except ImportError:
        print("Locust is not installed, exit.")
        exit(1)

    sys.argv[0] = 'locust'
    if len(sys.argv) == 1:
        sys.argv.extend(["-h"])

    if sys.argv[1] in ["-h", "--help", "-V", "--version"]:
        main()
        sys.exit(0)

    try:
        testcase_index = sys.argv.index('-f') + 1
        assert testcase_index < len(sys.argv)
    except (ValueError, AssertionError):
        print("Testcase file is not specified, exit.")
        sys.exit(1)

    testcase_file_path = sys.argv[testcase_index]
    sys.argv[testcase_index] = parse_locustfile(testcase_file_path)
    main()

def parse_locustfile(file_path):
    """ parse testcase file and return locustfile path.
        if file_path is a Python file, assume it is a locustfile
        if file_path is a YAML/JSON file, convert it to locustfile
    """
    if not os.path.isfile(file_path):
        print("file path invalid, exit.")
        sys.exit(1)

    file_suffix = os.path.splitext(file_path)[1]
    if file_suffix == ".py":
        locustfile_path = file_path
    elif file_suffix in ['.yaml', '.yml', '.json']:
        locustfile_path = gen_locustfile(file_path)
    else:
        # '' or other suffix
        print("file type should be YAML/JSON/Python, exit.")
        sys.exit(1)

    return locustfile_path

def gen_locustfile(testcase_file_path):
    """ generate locustfile from template.
    """
    locustfile_path = 'locustfile.py'
    with codecs.open('ate/locustfile_template', encoding='utf-8') as template:
        with codecs.open(locustfile_path, 'w', encoding='utf-8') as locustfile:
            template_content = template.read()
            template_content = template_content.replace("$HOST", "https://skypixel.com")
            template_content = template_content.replace("$TESTCASE_FILE", testcase_file_path)
            locustfile.write(template_content)

    return locustfile_path
