# encoding: utf-8

import copy

from httprunner import exceptions, logger, parser, utils
from httprunner.compat import OrderedDict


class Context(object):
    """ Manages context functions and variables.
        context has two levels, testcase and teststep.
    """
    def __init__(self, variables=None, functions=None):
        """ init Context with testcase variables and functions.
        """
        # testcase level context
        ## TESTCASE_SHARED_VARIABLES_MAPPING and TESTCASE_SHARED_FUNCTIONS_MAPPING will not change.
        self.TESTCASE_SHARED_VARIABLES_MAPPING = variables or OrderedDict()
        self.TESTCASE_SHARED_FUNCTIONS_MAPPING = functions or OrderedDict()

        # testcase level request, will not change
        self.TESTCASE_SHARED_REQUEST_MAPPING = {}

        self.evaluated_validators = []
        self.init_context_variables(level="testcase")

    def init_context_variables(self, level="testcase"):
        """ initialize testcase/teststep context

        Args:
            level (enum): "testcase" or "teststep"

        """
        if level == "testcase":
            # testcase level runtime context, will be updated with extracted variables in each teststep.
            self.testcase_runtime_variables_mapping = copy.deepcopy(self.TESTCASE_SHARED_VARIABLES_MAPPING)

        # teststep level context, will be altered in each teststep.
        # teststep config shall inherit from testcase configs,
        # but can not change testcase configs, that's why we use copy.deepcopy here.
        self.teststep_variables_mapping = copy.deepcopy(self.testcase_runtime_variables_mapping)

    def update_context_variables(self, variables, level):
        """ update context variables, with level specified.

        Args:
            variables (list/OrderedDict): testcase config block or teststep block
                [
                    {"TOKEN": "debugtalk"},
                    {"random": "${gen_random_string(5)}"},
                    {"json": {'name': 'user', 'password': '123456'}},
                    {"md5": "${gen_md5($TOKEN, $json, $random)}"}
                ]
                OrderDict({
                    "TOKEN": "debugtalk",
                    "random": "${gen_random_string(5)}",
                    "json": {'name': 'user', 'password': '123456'},
                    "md5": "${gen_md5($TOKEN, $json, $random)}"
                })
            level (enum): "testcase" or "teststep"

        """
        if isinstance(variables, list):
            variables = utils.convert_mappinglist_to_orderdict(variables)

        for variable_name, variable_value in variables.items():
            variable_eval_value = self.eval_content(variable_value)

            if level == "testcase":
                self.testcase_runtime_variables_mapping[variable_name] = variable_eval_value

            self.update_teststep_variables_mapping(variable_name, variable_eval_value)

    def eval_content(self, content):
        """ evaluate content recursively, take effect on each variable and function in content.
            content may be in any data structure, include dict, list, tuple, number, string, etc.
        """
        return parser.parse_data(
            content,
            self.teststep_variables_mapping,
            self.TESTCASE_SHARED_FUNCTIONS_MAPPING
        )

    def update_testcase_runtime_variables_mapping(self, variables):
        """ update testcase_runtime_variables_mapping with extracted vairables in teststep.

        Args:
            variables (OrderDict): extracted variables in teststep

        """
        for variable_name, variable_value in variables.items():
            self.testcase_runtime_variables_mapping[variable_name] = variable_value
            self.update_teststep_variables_mapping(variable_name, variable_value)

    def update_teststep_variables_mapping(self, variable_name, variable_value):
        """ bind and update testcase variables mapping
        """
        self.teststep_variables_mapping[variable_name] = variable_value

    def get_parsed_request(self, request_dict, level="teststep"):
        """ get parsed request with variables and functions.

        Args:
            request_dict (dict): request config mapping
            level (enum): "testcase" or "teststep"

        Returns:
            dict: parsed request dict

        """
        if level == "testcase":
            # testcase config request dict has been parsed in parse_tests
            self.TESTCASE_SHARED_REQUEST_MAPPING = copy.deepcopy(request_dict)
            return self.TESTCASE_SHARED_REQUEST_MAPPING

        else:
            # teststep
            return self.eval_content(
                utils.deep_update_dict(
                    copy.deepcopy(self.TESTCASE_SHARED_REQUEST_MAPPING),
                    request_dict
                )
            )

    def __eval_check_item(self, validator, resp_obj):
        """ evaluate check item in validator.

        Args:
            validator (dict): validator
                {"check": "status_code", "comparator": "eq", "expect": 201}
                {"check": "$resp_body_success", "comparator": "eq", "expect": True}
            resp_obj (object): requests.Response() object

        Returns:
            dict: validator info
                {
                    "check": "status_code",
                    "check_value": 200,
                    "expect": 201,
                    "comparator": "eq"
                }

        """
        check_item = validator["check"]
        # check_item should only be the following 5 formats:
        # 1, variable reference, e.g. $token
        # 2, function reference, e.g. ${is_status_code_200($status_code)}
        # 3, dict or list, maybe containing variable/function reference, e.g. {"var": "$abc"}
        # 4, string joined by delimiter. e.g. "status_code", "headers.content-type"
        # 5, regex string, e.g. "LB[\d]*(.*)RB[\d]*"

        if isinstance(check_item, (dict, list)) \
            or parser.extract_variables(check_item) \
            or parser.extract_functions(check_item):
            # format 1/2/3
            check_value = self.eval_content(check_item)
        else:
            # format 4/5
            check_value = resp_obj.extract_field(check_item)

        validator["check_value"] = check_value

        # expect_value should only be in 2 types:
        # 1, variable reference, e.g. $expect_status_code
        # 2, actual value, e.g. 200
        expect_value = self.eval_content(validator["expect"])
        validator["expect"] = expect_value
        validator["check_result"] = "unchecked"
        return validator

    def _do_validation(self, validator_dict):
        """ validate with functions

        Args:
            validator_dict (dict): validator dict
                {
                    "check": "status_code",
                    "check_value": 200,
                    "expect": 201,
                    "comparator": "eq"
                }

        """
        # TODO: move comparator uniform to init_test_suites
        comparator = utils.get_uniform_comparator(validator_dict["comparator"])
        validate_func = parser.get_mapping_function(comparator, self.TESTCASE_SHARED_FUNCTIONS_MAPPING)

        check_item = validator_dict["check"]
        check_value = validator_dict["check_value"]
        expect_value = validator_dict["expect"]

        if (check_value is None or expect_value is None) \
            and comparator not in ["is", "eq", "equals", "=="]:
            raise exceptions.ParamsError("Null value can only be compared with comparator: eq/equals/==")

        validate_msg = "validate: {} {} {}({})".format(
            check_item,
            comparator,
            expect_value,
            type(expect_value).__name__
        )

        try:
            validator_dict["check_result"] = "pass"
            validate_func(check_value, expect_value)
            validate_msg += "\t==> pass"
            logger.log_debug(validate_msg)
        except (AssertionError, TypeError):
            validate_msg += "\t==> fail"
            validate_msg += "\n{}({}) {} {}({})".format(
                check_value,
                type(check_value).__name__,
                comparator,
                expect_value,
                type(expect_value).__name__
            )
            logger.log_error(validate_msg)
            validator_dict["check_result"] = "fail"
            raise exceptions.ValidationFailure(validate_msg)

    def validate(self, validators, resp_obj):
        """ make validations
        """
        evaluated_validators = []
        if not validators:
            return evaluated_validators

        logger.log_info("start to validate.")
        validate_pass = True
        failures = []

        for validator in validators:
            # evaluate validators with context variable mapping.
            evaluated_validator = self.__eval_check_item(
                parser.parse_validator(validator),
                resp_obj
            )

            try:
                self._do_validation(evaluated_validator)
            except exceptions.ValidationFailure as ex:
                validate_pass = False
                failures.append(str(ex))

            evaluated_validators.append(evaluated_validator)

        if not validate_pass:
            failures_string = "\n".join([failure for failure in failures])
            raise exceptions.ValidationFailure(failures_string)

        return evaluated_validators
