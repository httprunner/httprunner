from httprunner import exceptions, logger, parser, utils


class SessionContext(object):
    """ HttpRunner session, store runtime variables.

    Examples:
        >>> variables = {"SECRET_KEY": "DebugTalk"}
        >>> context = SessionContext(variables)

        Equivalent to:
        >>> context = SessionContext()
        >>> context.update_session_variables(variables)

    """
    def __init__(self, variables=None):
        variables_mapping = utils.ensure_mapping_format(variables or {})
        self.session_variables_mapping = parser.parse_variables_mapping(variables_mapping)
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
        variables_mapping.update(self.session_variables_mapping)
        parsed_variables_mapping = parser.parse_variables_mapping(variables_mapping)

        self.test_variables_mapping = {}
        # priority: extracted variable > teststep variable
        self.test_variables_mapping.update(parsed_variables_mapping)
        self.test_variables_mapping.update(self.session_variables_mapping)

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
        return parser.parse_lazy_data(content, self.test_variables_mapping)

    def __eval_validator_check(self, check_item, resp_obj):
        """ evaluate check item in validator.

        Args:
            check_item: check_item should only be the following 5 formats:
                1, variable reference, e.g. $token
                2, function reference, e.g. ${is_status_code_200($status_code)}
                3, dict or list, maybe containing variable/function reference, e.g. {"var": "$abc"}
                4, string joined by delimiter. e.g. "status_code", "headers.content-type"
                5, regex string, e.g. "LB[\d]*(.*)RB[\d]*"

            resp_obj: response object

        """
        if isinstance(check_item, (dict, list)) \
            or isinstance(check_item, parser.LazyString):
            # format 1/2/3
            check_value = self.eval_content(check_item)
        else:
            # format 4/5
            check_value = resp_obj.extract_field(check_item)

        return check_value

    def __eval_validator_expect(self, expect_item):
        """ evaluate expect item in validator.

        Args:
            expect_item: expect_item should only be in 2 types:
                1, variable reference, e.g. $expect_status_code
                2, actual value, e.g. 200

        """
        expect_value = self.eval_content(expect_item)
        return expect_value

    def validate(self, validators, resp_obj):
        """ make validation with comparators
        """
        self.validation_results = []
        if not validators:
            return

        logger.log_debug("start to validate.")

        validate_pass = True
        failures = []

        for validator in validators:
            # validator should be LazyFunction object
            if not isinstance(validator, parser.LazyFunction):
                raise exceptions.ValidationFailure(
                    "validator should be parsed first: {}".format(validators))

            # evaluate validator args with context variable mapping.
            validator_args = validator.get_args()
            check_item, expect_item = validator_args
            check_value = self.__eval_validator_check(
                check_item,
                resp_obj
            )
            expect_value = self.__eval_validator_expect(expect_item)
            validator.update_args([check_value, expect_value])

            comparator = validator.func_name
            validator_dict = {
                "comparator": comparator,
                "check": check_item,
                "check_value": check_value,
                "expect": expect_item,
                "expect_value": expect_value
            }
            validate_msg = "\nvalidate: {} {} {}({})".format(
                check_item,
                comparator,
                expect_value,
                type(expect_value).__name__
            )

            try:
                validator.to_value(self.test_variables_mapping)
                validator_dict["check_result"] = "pass"
                validate_msg += "\t==> pass"
                logger.log_debug(validate_msg)
            except (AssertionError, TypeError):
                validate_pass = False
                validator_dict["check_result"] = "fail"
                validate_msg += "\t==> fail"
                validate_msg += "\n{}({}) {} {}({})".format(
                    check_value,
                    type(check_value).__name__,
                    comparator,
                    expect_value,
                    type(expect_value).__name__
                )
                logger.log_error(validate_msg)
                failures.append(validate_msg)

            self.validation_results.append(validator_dict)

            # restore validator args, in case of running multiple times
            validator.update_args(validator_args)

        if not validate_pass:
            failures_string = "\n".join([failure for failure in failures])
            raise exceptions.ValidationFailure(failures_string)
