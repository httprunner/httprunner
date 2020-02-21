# encoding: utf-8

import sys
import traceback

from httprunner import exceptions, logger, parser


class Validator(object):
    """Validate tests

    Attributes:
        validation_results (dict): store validation results,
            including validate_extractor and validate_script.

    """

    def __init__(self, session_context, resp_obj):
        """ initialize a Validator for each teststep (API request)

        Args:
            session_context: HttpRunner session context
            resp_obj: ResponseObject instance
        """
        self.session_context = session_context
        self.resp_obj = resp_obj
        self.validation_results = {}

    def __eval_validator_check(self, check_item):
        """ evaluate check item in validator.

        Args:
            check_item: check_item should only be the following 5 formats:
                1, variable reference, e.g. $token
                2, function reference, e.g. ${is_status_code_200($status_code)}
                3, dict or list, maybe containing variable/function reference, e.g. {"var": "$abc"}
                4, string joined by delimiter. e.g. "status_code", "headers.content-type"
                5, regex string, e.g. "LB[\d]*(.*)RB[\d]*"

        """
        if isinstance(check_item, (dict, list)) \
                or isinstance(check_item, parser.LazyString):
            # format 1/2/3
            check_value = self.session_context.eval_content(check_item)
        else:
            # format 4/5
            check_value = self.resp_obj.extract_field(check_item)

        return check_value

    def __eval_validator_expect(self, expect_item):
        """ evaluate expect item in validator.

        Args:
            expect_item: expect_item should only be in 2 types:
                1, variable reference, e.g. $expect_status_code
                2, actual value, e.g. 200

        """
        expect_value = self.session_context.eval_content(expect_item)
        return expect_value

    def validate_script(self, script):
        """ make validation with python script
        """
        result = {
            "validate_script": "<br/>".join(script),
            "check_result": "pass",
            "output": ""
        }

        script = "\n    ".join(script)
        code = """
# encoding: utf-8

def run_validate_script():
    {}
""".format(script)

        variables = {
            "status_code": self.resp_obj.status_code,
            "response_json": self.resp_obj.json,
            "response": self.resp_obj
        }
        variables.update(self.session_context.test_variables_mapping)
        variables.update(globals())

        try:
            exec(code, variables)
        except SyntaxError as ex:
            logger.log_warning("SyntaxError in python validate script: {}".format(ex))
            result["check_result"] = "fail"
            result["output"] = "<br/>".join([
                "ErrorMessage: {}".format(ex.msg),
                "ErrorLine: {}".format(ex.lineno),
                "ErrorText: {}".format(ex.text)
            ])
            return result

        try:
            # run python validate script
            variables["run_validate_script"]()
        except Exception as ex:
            logger.log_warning("run python validate script failed: {}".format(ex))
            result["check_result"] = "fail"

            _type, _value, _tb = sys.exc_info()

            _lineno = -1
            if _tb.tb_next:
                _lineno = _tb.tb_next.tb_lineno
                line_no = _lineno - 4
            elif len(traceback.extract_tb(_tb)) > 0:
                # filename, lineno, name, line
                _, _lineno, _, _ = traceback.extract_tb(_tb)[-1]
                line_no = _lineno - 4
            else:
                line_no = "N/A"

            result["output"] = "<br/>".join([
                "ErrorType: {}".format(_type.__name__),
                "ErrorLine: {}".format(line_no)
            ])

        return result

    def validate(self, validators):
        """ make validation with comparators
        """
        self.validation_results = {}
        if not validators:
            return

        logger.log_debug("start to validate.")

        validate_pass = True
        failures = []

        for validator in validators:

            if isinstance(validator, dict) and validator.get("type") == "python_script":
                script = self.session_context.eval_content(validator["script"])
                result = self.validate_script(script)
                if result["check_result"] == "fail":
                    validate_pass = False
                    failures.append(result["output"])

                self.validation_results["validate_script"] = result
                continue

            if "validate_extractor" not in self.validation_results:
                self.validation_results["validate_extractor"] = []

            # validator should be LazyFunction object
            if not isinstance(validator, parser.LazyFunction):
                raise exceptions.ValidationFailure(
                    "validator should be parsed first: {}".format(validators))

            # evaluate validator args with context variable mapping.
            validator_args = validator.get_args()
            check_item, expect_item = validator_args
            check_value = self.__eval_validator_check(check_item)
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
                validator.to_value(self.session_context.test_variables_mapping)
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

            self.validation_results["validate_extractor"].append(validator_dict)

            # restore validator args, in case of running multiple times
            validator.update_args(validator_args)

        if not validate_pass:
            failures_string = "\n".join([failure for failure in failures])
            raise exceptions.ValidationFailure(failures_string)
