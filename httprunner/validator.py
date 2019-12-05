# encoding: utf-8
import collections
import io
import json
import types

from httprunner import exceptions, logger, parser


###############################################################################
##   testcase validator utils
###############################################################################

def get_uniform_comparator(comparator):
    """ convert comparator alias to uniform name
    """
    if comparator in ["eq", "equals", "==", "is"]:
        return "equals"
    elif comparator in ["lt", "less_than"]:
        return "less_than"
    elif comparator in ["le", "less_than_or_equals"]:
        return "less_than_or_equals"
    elif comparator in ["gt", "greater_than"]:
        return "greater_than"
    elif comparator in ["ge", "greater_than_or_equals"]:
        return "greater_than_or_equals"
    elif comparator in ["ne", "not_equals"]:
        return "not_equals"
    elif comparator in ["str_eq", "string_equals"]:
        return "string_equals"
    elif comparator in ["len_eq", "length_equals", "count_eq"]:
        return "length_equals"
    elif comparator in ["len_gt", "count_gt", "length_greater_than", "count_greater_than"]:
        return "length_greater_than"
    elif comparator in ["len_ge", "count_ge", "length_greater_than_or_equals",
                        "count_greater_than_or_equals"]:
        return "length_greater_than_or_equals"
    elif comparator in ["len_lt", "count_lt", "length_less_than", "count_less_than"]:
        return "length_less_than"
    elif comparator in ["len_le", "count_le", "length_less_than_or_equals",
                        "count_less_than_or_equals"]:
        return "length_less_than_or_equals"
    else:
        return comparator


def uniform_validator(validator):
    """ unify validator

    Args:
        validator (dict): validator maybe in two formats:

            format1: this is kept for compatiblity with the previous versions.
                {"check": "status_code", "comparator": "eq", "expect": 201}
                {"check": "$resp_body_success", "comparator": "eq", "expect": True}
            format2: recommended new version, {comparator: [check_item, expected_value]}
                {'eq': ['status_code', 201]}
                {'eq': ['$resp_body_success', True]}

    Returns
        dict: validator info

            {
                "check": "status_code",
                "expect": 201,
                "comparator": "equals"
            }

    """
    if not isinstance(validator, dict):
        raise exceptions.ParamsError("invalid validator: {}".format(validator))

    if "check" in validator and "expect" in validator:
        # format1
        check_item = validator["check"]
        expect_value = validator["expect"]
        comparator = validator.get("comparator", "eq")

    elif len(validator) == 1:
        # format2
        comparator = list(validator.keys())[0]
        compare_values = validator[comparator]

        if not isinstance(compare_values, list) or len(compare_values) != 2:
            raise exceptions.ParamsError("invalid validator: {}".format(validator))

        check_item, expect_value = compare_values

    else:
        raise exceptions.ParamsError("invalid validator: {}".format(validator))

    # uniform comparator, e.g. lt => less_than, eq => equals
    comparator = get_uniform_comparator(comparator)

    return {
        "check": check_item,
        "expect": expect_value,
        "comparator": comparator
    }


def _convert_validators_to_mapping(validators):
    """ convert validators list to mapping.

    Args:
        validators (list): validators in list

    Returns:
        dict: validators mapping, use (check, comparator) as key.

    Examples:
        >>> validators = [
            {"check": "v1", "expect": 201, "comparator": "eq"},
            {"check": {"b": 1}, "expect": 200, "comparator": "eq"}
        ]
        >>> print(_convert_validators_to_mapping(validators))
        {
            ("v1", "eq"): {"check": "v1", "expect": 201, "comparator": "eq"},
            ('{"b": 1}', "eq"): {"check": {"b": 1}, "expect": 200, "comparator": "eq"}
        }

    """
    validators_mapping = {}

    for validator in validators:
        if not isinstance(validator["check"], collections.Hashable):
            check = json.dumps(validator["check"])
        else:
            check = validator["check"]

        key = (check, validator["comparator"])
        validators_mapping[key] = validator

    return validators_mapping


def extend_validators(raw_validators, override_validators):
    """ extend raw_validators with override_validators.
        override_validators will merge and override raw_validators.

    Args:
        raw_validators (dict):
        override_validators (dict):

    Returns:
        list: extended validators

    Examples:
        >>> raw_validators = [{'eq': ['v1', 200]}, {"check": "s2", "expect": 16, "comparator": "len_eq"}]
        >>> override_validators = [{"check": "v1", "expect": 201}, {'len_eq': ['s3', 12]}]
        >>> extend_validators(raw_validators, override_validators)
            [
                {"check": "v1", "expect": 201, "comparator": "eq"},
                {"check": "s2", "expect": 16, "comparator": "len_eq"},
                {"check": "s3", "expect": 12, "comparator": "len_eq"}
            ]

    """

    if not raw_validators:
        return override_validators

    elif not override_validators:
        return raw_validators

    else:
        def_validators_mapping = _convert_validators_to_mapping(raw_validators)
        ref_validators_mapping = _convert_validators_to_mapping(override_validators)

        def_validators_mapping.update(ref_validators_mapping)
        return list(def_validators_mapping.values())


###############################################################################
##   validate varibles and functions
###############################################################################


def is_function(item):
    """ Takes item object, returns True if it is a function.
    """
    return isinstance(item, types.FunctionType)


def is_variable(tup):
    """ Takes (name, object) tuple, returns True if it is a variable.
    """
    name, item = tup
    if callable(item):
        # function or class
        return False

    if isinstance(item, types.ModuleType):
        # imported module
        return False

    if name.startswith("_"):
        # private property
        return False

    return True


def validate_json_file(file_list):
    """ validate JSON testcase format
    """
    for json_file in set(file_list):
        if not json_file.endswith(".json"):
            logger.log_warning("Only JSON file format can be validated, skip: {}".format(json_file))
            continue

        logger.color_print("Start to validate JSON file: {}".format(json_file), "GREEN")

        with io.open(json_file) as stream:
            try:
                json.load(stream)
            except ValueError as e:
                raise SystemExit(e)

        print("OK")


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
        validator_dict = {
            "validate_script": "<br/>".join(script),
            "check_result": "fail",
            "exception": ""
        }

        script = "\n    ".join(script)
        code = """
# encoding: utf-8

try:
    {}
except Exception as ex:
    import traceback
    import sys
    _type, _value, _tb = sys.exc_info()
    # filename, lineno, name, line
    _, _lineno, _, line_content = traceback.extract_tb(_tb, 1)[0]

    line_no = _lineno - 4

    c_exception = _type.__name__ + "\\n"
    c_exception += "\\tError line number: " + str(line_no) + "\\n"
    c_exception += "\\tError line content: " + str(line_content) + "\\n"

    if _value.args:
        c_exception += "\\tError description: " + str(_value)
    else:
        c_exception += "\\tError description: " + _type.__name__

    raise _type(c_exception)
""".format(script)
        variables = {
            "status_code": self.resp_obj.status_code,
            "response_json": self.resp_obj.json,
            "response": self.resp_obj
        }
        variables.update(self.session_context.test_variables_mapping)

        try:
            code = compile(code, '<string>', 'exec')
            exec(code, variables)
            validator_dict["check_result"] = "pass"
            return validator_dict, ""
        except Exception as ex:
            validator_dict["check_result"] = "fail"
            validator_dict["exception"] = "<br/>".join(str(ex).splitlines())
            return validator_dict, str(ex)

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
                validator_dict, ex = self.validate_script(validator["script"])
                if ex:
                    validate_pass = False
                    failures.append(ex)

                self.validation_results["validate_script"] = validator_dict
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
