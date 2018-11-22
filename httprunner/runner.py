# encoding: utf-8

from unittest.case import SkipTest

from httprunner import exceptions, logger, response, utils
from httprunner.client import HttpSession
from httprunner.context import SessionContext


class Runner(object):
    """ Running testcases.

    Examples:
        >>> functions={...}
        >>> config = {
                "name": "XXXX",
                "base_url": "http://127.0.0.1",
                "verify": False
            }
        >>> runner = Runner(config, functions)

        >>> teststep = {
                "name": "teststep description",
                "variables": [],        # optional
                "request": {
                    "url": "http://127.0.0.1:5000/api/users/1000",
                    "method": "GET"
                }
            }
        >>> runner.run_test(teststep)

    """

    def __init__(self, config, functions, http_client_session=None):
        """ run testcase or testsuite.

        Args:
            config (dict): testcase/testsuite config dict

                {
                    "name": "ABC",
                    "variables": {},
                    "setup_hooks", [],
                    "teardown_hooks", []
                }

            http_client_session (instance): requests.Session(), or locust.client.Session() instance.

        """
        base_url = config.get("base_url")
        self.verify = config.get("verify", True)
        self.output = config.get("output", [])
        self.functions = functions
        self.evaluated_validators = []

        # testcase setup hooks
        testcase_setup_hooks = config.get("setup_hooks", [])
        # testcase teardown hooks
        self.testcase_teardown_hooks = config.get("teardown_hooks", [])

        self.http_client_session = http_client_session or HttpSession(base_url)
        self.session_context = SessionContext(self.functions)

        if testcase_setup_hooks:
            self.do_hook_actions(testcase_setup_hooks)

    def __del__(self):
        if self.testcase_teardown_hooks:
            self.do_hook_actions(self.testcase_teardown_hooks)

    def _handle_skip_feature(self, teststep_dict):
        """ handle skip feature for teststep
            - skip: skip current test unconditionally
            - skipIf: skip current test if condition is true
            - skipUnless: skip current test unless condition is true

        Args:
            teststep_dict (dict): teststep info

        Raises:
            SkipTest: skip teststep

        """
        # TODO: move skip to initialize
        skip_reason = None

        if "skip" in teststep_dict:
            skip_reason = teststep_dict["skip"]

        elif "skipIf" in teststep_dict:
            skip_if_condition = teststep_dict["skipIf"]
            if self.session_context.eval_content(skip_if_condition):
                skip_reason = "{} evaluate to True".format(skip_if_condition)

        elif "skipUnless" in teststep_dict:
            skip_unless_condition = teststep_dict["skipUnless"]
            if not self.session_context.eval_content(skip_unless_condition):
                skip_reason = "{} evaluate to False".format(skip_unless_condition)

        if skip_reason:
            raise SkipTest(skip_reason)

    def do_hook_actions(self, actions):
        for action in actions:
            logger.log_debug("call hook: {}".format(action))
            # TODO: check hook function if valid
            self.session_context.eval_content(action)

    def _run_teststep(self, teststep_dict):
        """ run single teststep.

        Args:
            teststep_dict (dict): teststep info
                {
                    "name": "teststep description",
                    "skip": "skip this test unconditionally",
                    "times": 3,
                    "variables": [],            # optional, override
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

        Raises:
            exceptions.ParamsError
            exceptions.ValidationFailure
            exceptions.ExtractFailure

        """
        # check skip
        self._handle_skip_feature(teststep_dict)

        # prepare
        teststep_dict = utils.lower_test_dict_keys(teststep_dict)
        teststep_variables = teststep_dict.get("variables", {})
        self.session_context.init_teststep_variables(teststep_variables)

        # parse teststep request
        raw_request = teststep_dict.get('request', {})
        parsed_teststep_request = self.session_context.eval_content(raw_request)
        self.session_context.update_teststep_variables("request", parsed_teststep_request)

        # setup hooks
        setup_hooks = teststep_dict.get("setup_hooks", [])
        setup_hooks.insert(0, "${setup_hook_prepare_kwargs($request)}")
        self.do_hook_actions(setup_hooks)

        try:
            url = parsed_teststep_request.pop('url')
            method = parsed_teststep_request.pop('method')
            parsed_teststep_request.setdefault("verify", self.verify)
            group_name = parsed_teststep_request.pop("group", None)
        except KeyError:
            raise exceptions.ParamsError("URL or METHOD missed!")

        # TODO: move method validation to json schema
        valid_methods = ["GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"]
        if method.upper() not in valid_methods:
            err_msg = u"Invalid HTTP method! => {}\n".format(method)
            err_msg += "Available HTTP methods: {}".format("/".join(valid_methods))
            logger.log_error(err_msg)
            raise exceptions.ParamsError(err_msg)

        logger.log_info("{method} {url}".format(method=method, url=url))
        logger.log_debug("request kwargs(raw): {kwargs}".format(kwargs=parsed_teststep_request))

        # request
        resp = self.http_client_session.request(
            method,
            url,
            name=group_name,
            **parsed_teststep_request
        )
        resp_obj = response.ResponseObject(resp)

        # teardown hooks
        teardown_hooks = teststep_dict.get("teardown_hooks", [])
        if teardown_hooks:
            logger.log_info("start to run teardown hooks")
            self.session_context.update_teststep_variables("response", resp_obj)
            self.do_hook_actions(teardown_hooks)

        # extract
        extractors = teststep_dict.get("extract", [])
        extracted_variables_mapping = resp_obj.extract_response(extractors)
        self.session_context.update_seesion_variables(extracted_variables_mapping)

        # validate
        validators = teststep_dict.get("validate", [])
        try:
            self.evaluated_validators = self.session_context.validate(validators, resp_obj)
        except (exceptions.ParamsError, exceptions.ValidationFailure, exceptions.ExtractFailure):
            # log request
            err_req_msg = "request: \n"
            err_req_msg += "headers: {}\n".format(parsed_teststep_request.pop("headers", {}))
            for k, v in parsed_teststep_request.items():
                err_req_msg += "{}: {}\n".format(k, repr(v))
            logger.log_error(err_req_msg)

            # log response
            err_resp_msg = "response: \n"
            err_resp_msg += "status_code: {}\n".format(resp_obj.status_code)
            err_resp_msg += "headers: {}\n".format(resp_obj.headers)
            err_resp_msg += "body: {}\n".format(repr(resp_obj.text))
            logger.log_error(err_resp_msg)

            raise

    def _run_testcase(self, testcase_dict):
        """ run single testcase.
        """
        config = testcase_dict.get("config", {})
        test_runner = Runner(config, self.functions, self.http_client_session)

        teststeps = testcase_dict.get("teststeps", [])
        for index, teststep_dict in enumerate(teststeps):
            test_runner.run_test(teststep_dict)

        self.session_context.update_seesion_variables(test_runner.extract_sessions())

    def run_test(self, teststep_dict):
        """ run single teststep of testcase.
            teststep_dict may be in 3 types.

        Args:
            teststep_dict (dict):

                # teststep
                {
                    "name": "teststep description",
                    "variables": [],        # optional
                    "request": {
                        "url": "http://127.0.0.1:5000/api/users/1000",
                        "method": "GET"
                    }
                }

                # embeded testcase
                {
                    "config": {...},
                    "teststeps": [
                        {...},
                        {...}
                    ]
                }

                # TODO: function
                {
                    "name": "exec function",
                    "function": "${func()}"
                }

        """
        if "config" in teststep_dict:
            self._run_testcase(teststep_dict)
        else:
            # api
            self._run_teststep(teststep_dict)

    def extract_sessions(self):
        """
        """
        return self.extract_output(self.output)

    def extract_output(self, output_variables_list):
        """ extract output variables
        """
        variables_mapping = self.session_context.session_variables_mapping

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
