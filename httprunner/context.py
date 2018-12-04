from httprunner import exceptions, logger, parser, utils


class SessionContext(object):
    """ HttpRunner session, store runtime variables.

    Examples:
        >>> functions={...}
        >>> variables = {"SECRET_KEY": "DebugTalk"}
        >>> context = SessionContext(functions, variables)

        Equivalent to:
        >>> context = SessionContext(functions)
        >>> context.update_session_variables(variables)

    """
    def __init__(self, functions, variables=None):
        self.session_variables_mapping = utils.ensure_mapping_format(variables or {})
        self.FUNCTIONS_MAPPING = functions
        self.init_test_variables()
        self.validation_results = []

    def init_test_variables(self, variables_mapping=None):
        """ init test variables, called when each test(api) starts.
            variables_mapping will be evaluated first.

        Args:
            variables_mapping (dict)
                {
                    "random": "${gen_random_string(5)}",
                    "authorization": "${gen_md5($TOKEN, $data, $random)}",
                    "data": '{"name": "user", "password": "123456"}',
                    "TOKEN": "debugtalk",
                }

        """
        variables_mapping = variables_mapping or {}
        variables_mapping = utils.ensure_mapping_format(variables_mapping)

        self.test_variables_mapping = {}
        # priority: extracted variable > teststep variable
        self.test_variables_mapping.update(variables_mapping)
        self.test_variables_mapping.update(self.session_variables_mapping)

        for variable_name, variable_value in variables_mapping.items():
            variable_value = self.eval_content(variable_value)
            self.update_test_variables(variable_name, variable_value)

    def update_test_variables(self, variable_name, variable_value):
        """ update test variables, these variables are only valid in the current test.
        """
        self.test_variables_mapping[variable_name] = variable_value

    def update_session_variables(self, variables_mapping):
        """ update session with extracted variables mapping.
            these variables are valid in the whole running session.
        """
        variables_mapping = utils.ensure_mapping_format(variables_mapping)
        self.session_variables_mapping.update(variables_mapping)
        self.test_variables_mapping.update(self.session_variables_mapping)

    def eval_content(self, content):
        """ evaluate content recursively, take effect on each variable and function in content.
            content may be in any data structure, include dict, list, tuple, number, string, etc.
        """
        return parser.parse_data(
            content,
            self.test_variables_mapping,
            self.FUNCTIONS_MAPPING
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
        validate_func = parser.get_mapping_function(comparator, self.FUNCTIONS_MAPPING)

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
        if not validators:
            return

        logger.log_debug("start to validate.")

        self.validation_results = []
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

            self.validation_results.append(evaluated_validator)

        if not validate_pass:
            failures_string = "\n".join([failure for failure in failures])
            raise exceptions.ValidationFailure(failures_string)
