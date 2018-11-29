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

        >>> test_dict = {
                "name": "test description",
                "variables": [],        # optional
                "request": {
                    "url": "http://127.0.0.1:5000/api/users/1000",
                    "method": "GET"
                }
            }
        >>> runner.run_test(test_dict)

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
        self.validation_results = []

        # testcase setup hooks
        testcase_setup_hooks = config.get("setup_hooks", [])
        # testcase teardown hooks
        self.testcase_teardown_hooks = config.get("teardown_hooks", [])

        self.http_client_session = http_client_session or HttpSession(base_url)
        self.session_context = SessionContext(self.functions)

        if testcase_setup_hooks:
            self.do_hook_actions(testcase_setup_hooks, "setup")

    def __del__(self):
        if self.testcase_teardown_hooks:
            self.do_hook_actions(self.testcase_teardown_hooks, "teardown")

    def __clear_test_data(self):
        """ clear request and response data
        """
        if not isinstance(self.http_client_session, HttpSession):
            return

        self.validation_results = []
        self.http_client_session.init_meta_data()

    def __get_test_data(self):
        """ get request/response data and validate results
        """
        if not isinstance(self.http_client_session, HttpSession):
            return

        meta_data = self.http_client_session.meta_data
        meta_data["validators"] = self.validation_results
        return meta_data

    def _handle_skip_feature(self, test_dict):
        """ handle skip feature for test
            - skip: skip current test unconditionally
            - skipIf: skip current test if condition is true
            - skipUnless: skip current test unless condition is true

        Args:
            test_dict (dict): test info

        Raises:
            SkipTest: skip test

        """
        # TODO: move skip to initialize
        skip_reason = None

        if "skip" in test_dict:
            skip_reason = test_dict["skip"]

        elif "skipIf" in test_dict:
            skip_if_condition = test_dict["skipIf"]
            if self.session_context.eval_content(skip_if_condition):
                skip_reason = "{} evaluate to True".format(skip_if_condition)

        elif "skipUnless" in test_dict:
            skip_unless_condition = test_dict["skipUnless"]
            if not self.session_context.eval_content(skip_unless_condition):
                skip_reason = "{} evaluate to False".format(skip_unless_condition)

        if skip_reason:
            raise SkipTest(skip_reason)

    def do_hook_actions(self, actions, hook_type):
        """ call hook actions.

        Args:
            actions (str/dict): actions maybe in two format.

                format1: only call hook functions.
                    ${func()}
                format2: assignment, the value returned by hook function will be assigned to variable.
                    {"var": "${func()}"}

            hook_type (enum): setup/teardown

        """
        logger.log_debug("call {} hook actions.".format(hook_type))
        for action in actions:

            if isinstance(action, dict) and len(action) == 1:
                # {"var": "${func()}"}
                var_name, hook_content = list(action.items())[0]
                logger.log_debug("assignment with hook: {} = {}".format(var_name, hook_content))
                self.session_context.update_test_variables(
                    var_name,
                    self.session_context.eval_content(hook_content)
                )
            else:
                logger.log_debug("call hook function: {}".format(action))
                # TODO: check hook function if valid
                self.session_context.eval_content(action)

    def _run_test(self, test_dict):
        """ run single teststep.

        Args:
            test_dict (dict): teststep info
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
                        "json": {"name": "user", "password": "123456"}
                    },
                    "extract": {},              # optional
                    "validate": [],             # optional
                    "setup_hooks": [],          # optional
                    "teardown_hooks": []        # optional
                }

        Raises:
            exceptions.ParamsError
            exceptions.ValidationFailure
            exceptions.ExtractFailure

        """
        # clear meta data first to ensure independence for each test
        self.__clear_test_data()

        # check skip
        self._handle_skip_feature(test_dict)

        # prepare
        test_dict = utils.lower_test_dict_keys(test_dict)
        test_variables = test_dict.get("variables", {})
        self.session_context.init_test_variables(test_variables)

        # parse test request
        raw_request = test_dict.get('request', {})
        parsed_test_request = self.session_context.eval_content(raw_request)
        self.session_context.update_test_variables("request", parsed_test_request)

        # setup hooks
        setup_hooks = test_dict.get("setup_hooks", [])
        setup_hooks.insert(0, "${setup_hook_prepare_kwargs($request)}")
        self.do_hook_actions(setup_hooks, "setup")

        try:
            url = parsed_test_request.pop('url')
            method = parsed_test_request.pop('method')
            parsed_test_request.setdefault("verify", self.verify)
            group_name = parsed_test_request.pop("group", None)
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
        logger.log_debug("request kwargs(raw): {kwargs}".format(kwargs=parsed_test_request))

        # request
        resp = self.http_client_session.request(
            method,
            url,
            name=group_name,
            **parsed_test_request
        )
        resp_obj = response.ResponseObject(resp)

        # teardown hooks
        teardown_hooks = test_dict.get("teardown_hooks", [])
        if teardown_hooks:
            logger.log_info("start to run teardown hooks")
            self.session_context.update_test_variables("response", resp_obj)
            self.do_hook_actions(teardown_hooks, "teardown")

        # extract
        extractors = test_dict.get("extract", {})
        extracted_variables_mapping = resp_obj.extract_response(extractors)
        self.session_context.update_seesion_variables(extracted_variables_mapping)

        # validate
        validators = test_dict.get("validate", [])
        try:
            self.session_context.validate(validators, resp_obj)

        except (exceptions.ParamsError, exceptions.ValidationFailure, exceptions.ExtractFailure):
            # log request
            err_req_msg = "request: \n"
            err_req_msg += "headers: {}\n".format(parsed_test_request.pop("headers", {}))
            for k, v in parsed_test_request.items():
                err_req_msg += "{}: {}\n".format(k, repr(v))
            logger.log_error(err_req_msg)

            # log response
            err_resp_msg = "response: \n"
            err_resp_msg += "status_code: {}\n".format(resp_obj.status_code)
            err_resp_msg += "headers: {}\n".format(resp_obj.headers)
            err_resp_msg += "body: {}\n".format(repr(resp_obj.text))
            logger.log_error(err_resp_msg)

            raise

        finally:
            self.validation_results = self.session_context.validation_results

    def _run_testcase(self, testcase_dict):
        """ run single testcase.
        """
        self.meta_datas = []
        config = testcase_dict.get("config", {})
        test_runner = Runner(config, self.functions, self.http_client_session)

        tests = testcase_dict.get("tests", [])

        try:
            for index, test_dict in enumerate(tests):
                test_runner.run_test(test_dict)
                _meta_datas = test_runner.meta_datas
                self.meta_datas.append(_meta_datas)

            self.session_context.update_seesion_variables(test_runner.extract_sessions())

        except Exception as ex:
            meta_datas = test_runner.meta_datas
            self.meta_datas.append(meta_datas)
            raise

    def run_test(self, test_dict):
        """ run single teststep of testcase.
            test_dict may be in 3 types.

        Args:
            test_dict (dict):

                # teststep
                {
                    "name": "teststep description",
                    "variables": [],        # optional
                    "request": {
                        "url": "http://127.0.0.1:5000/api/users/1000",
                        "method": "GET"
                    }
                }

                # nested testcase
                {
                    "config": {...},
                    "tests": [
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
        self.meta_datas = None
        if "config" in test_dict:
            # nested testcase
            self._run_testcase(test_dict)
        else:
            # api
            try:
                self._run_test(test_dict)
            except Exception:
                raise
            finally:
                self.meta_datas = self.__get_test_data()

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
