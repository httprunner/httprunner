# encoding: utf-8

from unittest.case import SkipTest

from httprunner import exception, logger, response, utils
from httprunner.client import HttpSession
from httprunner.context import Context


class Runner(object):

    def __init__(self, config_dict=None, http_client_session=None):
        self.http_client_session = http_client_session
        self.evaluated_validators = []
        self.context = Context()

        config_dict = config_dict or {}
        self.init_config(config_dict, "testset")

        # testset setup hooks
        testset_setup_hooks = config_dict.pop("setup_hooks", [])
        if testset_setup_hooks:
            self.do_hook_actions(testset_setup_hooks)

        # testset teardown hooks
        self.testset_teardown_hooks = config_dict.pop("teardown_hooks", [])

    def __del__(self):
        if self.testset_teardown_hooks:
            self.do_hook_actions(self.testset_teardown_hooks)

    def init_config(self, config_dict, level):
        """ create/update context variables binds
        @param (dict) config_dict
        @param (str) level, "testset" or "testcase"
        testset:
            {
                "name": "smoke testset",
                "path": "tests/data/demo_testset_variables.yml",
                "requires": [],         # optional
                "function_binds": {},   # optional
                "import_module_items": [],  # optional
                "variables": [],   # optional
                "request": {
                    "base_url": "http://127.0.0.1:5000",
                    "headers": {
                        "User-Agent": "iOS/2.8.3"
                    }
                }
            }
        testcase:
            {
                "name": "testcase description",
                "requires": [],         # optional
                "function_binds": {},   # optional
                "import_module_items": [],  # optional
                "variables": [],   # optional
                "request": {
                    "url": "/api/get-token",
                    "method": "POST",
                    "headers": {
                        "Content-Type": "application/json"
                    }
                },
                "json": {
                    "sign": "f1219719911caae89ccc301679857ebfda115ca2"
                }
            }
        @param (str) context level, testcase or testset
        """
        # convert keys in request headers to lowercase
        config_dict = utils.lower_config_dict_key(config_dict)

        self.context.init_context(level)
        self.context.config_context(config_dict, level)

        request_config = config_dict.get('request', {})
        parsed_request = self.context.get_parsed_request(request_config, level)

        base_url = parsed_request.pop("base_url", None)
        self.http_client_session = self.http_client_session or HttpSession(base_url)

        return parsed_request

    def _handle_skip_feature(self, testcase_dict):
        """ handle skip feature for testcase
            - skip: skip current test unconditionally
            - skipIf: skip current test if condition is true
            - skipUnless: skip current test unless condition is true
        """
        skip_reason = None

        if "skip" in testcase_dict:
            skip_reason = testcase_dict["skip"]

        elif "skipIf" in testcase_dict:
            skip_if_condition = testcase_dict["skipIf"]
            if self.context.eval_content(skip_if_condition):
                skip_reason = "{} evaluate to True".format(skip_if_condition)

        elif "skipUnless" in testcase_dict:
            skip_unless_condition = testcase_dict["skipUnless"]
            if not self.context.eval_content(skip_unless_condition):
                skip_reason = "{} evaluate to False".format(skip_unless_condition)

        if skip_reason:
            raise SkipTest(skip_reason)

    def do_hook_actions(self, actions):
        for action in actions:
            logger.log_debug("call hook: {}".format(action))
            self.context.eval_content(action)

    def run_test(self, testcase_dict):
        """ run single testcase.
        @param (dict) testcase_dict
            {
                "name": "testcase description",
                "skip": "skip this test unconditionally",
                "times": 3,
                "requires": [],         # optional, override
                "function_binds": {},   # optional, override
                "variables": [],        # optional, override
                "request": {
                    "url": "http://127.0.0.1:5000/api/users/1000",
                    "method": "POST",
                    "headers": {
                        "Content-Type": "application/json",
                        "authorization": "$authorization",
                        "random": "$random"
                    },
                    "body": '{"name": "user", "password": "123456"}'
                },
                "extract": [],              # optional
                "validate": [],             # optional
                "setup_hooks": [],          # optional
                "teardown_hooks": []        # optional
            }
        @return True or raise exception during test
        """
        # check skip
        self._handle_skip_feature(testcase_dict)

        # prepare
        parsed_request = self.init_config(testcase_dict, level="testcase")
        self.context.bind_testcase_variable("request", parsed_request)

        # setup hooks
        setup_hooks = testcase_dict.get("setup_hooks", [])
        setup_hooks.insert(0, "${setup_hook_prepare_kwargs($request)}")
        self.do_hook_actions(setup_hooks)

        try:
            url = parsed_request.pop('url')
            method = parsed_request.pop('method')
            group_name = parsed_request.pop("group", None)
        except KeyError:
            raise exception.ParamsError("URL or METHOD missed!")

        logger.log_info("{method} {url}".format(method=method, url=url))
        logger.log_debug("request kwargs(raw): {kwargs}".format(kwargs=parsed_request))

        # request
        resp = self.http_client_session.request(
            method,
            url,
            name=group_name,
            **parsed_request
        )
        resp_obj = response.ResponseObject(resp)

        # teardown hooks
        teardown_hooks = testcase_dict.get("teardown_hooks", [])
        if teardown_hooks:
            self.context.bind_testcase_variable("response", resp_obj)
            self.do_hook_actions(teardown_hooks)

        # extract
        extractors = testcase_dict.get("extract", []) or testcase_dict.get("extractors", [])
        extracted_variables_mapping = resp_obj.extract_response(extractors)
        self.context.bind_extracted_variables(extracted_variables_mapping)

        # validate
        validators = testcase_dict.get("validate", []) or testcase_dict.get("validators", [])
        try:
            self.evaluated_validators = self.context.eval_validators(validators, resp_obj)
            self.context.validate(self.evaluated_validators)
        except (exception.ParamsError, exception.ResponseError, \
            exception.ValidationError, exception.ParseResponseError):
            # log request
            err_req_msg = "request: \n"
            err_req_msg += "headers: {}\n".format(parsed_request.pop("headers", {}))
            for k, v in parsed_request.items():
                err_req_msg += "{}: {}\n".format(k, v)
            logger.log_error(err_req_msg)

            # log response
            err_resp_msg = "response: \n"
            err_resp_msg += "status_code: {}\n".format(resp_obj.status_code)
            err_resp_msg += "headers: {}\n".format(resp_obj.headers)
            err_resp_msg += "content: {}\n".format(resp_obj.content)
            logger.log_error(err_resp_msg)

            raise

    def extract_output(self, output_variables_list):
        """ extract output variables
        """
        variables_mapping = self.context.testcase_variables_mapping

        output = {}
        for variable in output_variables_list:
            if variable not in variables_mapping:
                logger.log_warning(
                    "variable '{}' can not be found in variables mapping, failed to output!"\
                        .format(variable)
                )
                continue

            output[variable] = variables_mapping[variable]

        return output
