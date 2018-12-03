# encoding: utf-8

import ast
import os
import re

from httprunner import exceptions, utils
from httprunner.compat import basestring, builtin_str, numeric_types, str

variable_regexp = r"\$([\w_]+)"
function_regexp = r"\$\{([\w_]+\([\$\w\.\-/_ =,]*\))\}"
function_regexp_compile = re.compile(r"^([\w_]+)\(([\$\w\.\-/_ =,]*)\)$")


def parse_string_value(str_value):
    """ parse string to number if possible
    e.g. "123" => 123
         "12.2" => 12.3
         "abc" => "abc"
         "$var" => "$var"
    """
    try:
        return ast.literal_eval(str_value)
    except ValueError:
        return str_value
    except SyntaxError:
        # e.g. $var, ${func}
        return str_value


def extract_variables(content):
    """ extract all variable names from content, which is in format $variable

    Args:
        content (str): string content

    Returns:
        list: variables list extracted from string content

    Examples:
        >>> extract_variables("$variable")
        ["variable"]

        >>> extract_variables("/blog/$postid")
        ["postid"]

        >>> extract_variables("/$var1/$var2")
        ["var1", "var2"]

        >>> extract_variables("abc")
        []

    """
    # TODO: change variable notation from $var to {{var}}
    try:
        return re.findall(variable_regexp, content)
    except TypeError:
        return []


def extract_functions(content):
    """ extract all functions from string content, which are in format ${fun()}

    Args:
        content (str): string content

    Returns:
        list: functions list extracted from string content

    Examples:
        >>> extract_functions("${func(5)}")
        ["func(5)"]

        >>> extract_functions("${func(a=1, b=2)}")
        ["func(a=1, b=2)"]

        >>> extract_functions("/api/1000?_t=${get_timestamp()}")
        ["get_timestamp()"]

        >>> extract_functions("/api/${add(1, 2)}")
        ["add(1, 2)"]

        >>> extract_functions("/api/${add(1, 2)}?_t=${get_timestamp()}")
        ["add(1, 2)", "get_timestamp()"]

    """
    try:
        return re.findall(function_regexp, content)
    except TypeError:
        return []


def parse_function(content):
    """ parse function name and args from string content.

    Args:
        content (str): string content

    Returns:
        dict: function meta dict

            {
                "func_name": "xxx",
                "args": [],
                "kwargs": {}
            }

    Examples:
        >>> parse_function("func()")
        {'func_name': 'func', 'args': [], 'kwargs': {}}

        >>> parse_function("func(5)")
        {'func_name': 'func', 'args': [5], 'kwargs': {}}

        >>> parse_function("func(1, 2)")
        {'func_name': 'func', 'args': [1, 2], 'kwargs': {}}

        >>> parse_function("func(a=1, b=2)")
        {'func_name': 'func', 'args': [], 'kwargs': {'a': 1, 'b': 2}}

        >>> parse_function("func(1, 2, a=3, b=4)")
        {'func_name': 'func', 'args': [1, 2], 'kwargs': {'a':3, 'b':4}}

    """
    matched = function_regexp_compile.match(content)
    if not matched:
        raise exceptions.FunctionNotFound("{} not found!".format(content))

    function_meta = {
        "func_name": matched.group(1),
        "args": [],
        "kwargs": {}
    }

    args_str = matched.group(2).strip()
    if args_str == "":
        return function_meta

    args_list = args_str.split(',')
    for arg in args_list:
        arg = arg.strip()
        if '=' in arg:
            key, value = arg.split('=')
            function_meta["kwargs"][key.strip()] = parse_string_value(value.strip())
        else:
            function_meta["args"].append(parse_string_value(arg))

    return function_meta


def parse_validator(validator):
    """ parse validator

    Args:
        validator (dict): validator maybe in two formats:

            format1: this is kept for compatiblity with the previous versions.
                {"check": "status_code", "comparator": "eq", "expect": 201}
                {"check": "$resp_body_success", "comparator": "eq", "expect": True}
            format2: recommended new version
                {'eq': ['status_code', 201]}
                {'eq': ['$resp_body_success', True]}

    Returns
        dict: validator info

            {
                "check": "status_code",
                "expect": 201,
                "comparator": "eq"
            }

    """
    if not isinstance(validator, dict):
        raise exceptions.ParamsError("invalid validator: {}".format(validator))

    if "check" in validator and len(validator) > 1:
        # format1
        check_item = validator.get("check")

        if "expect" in validator:
            expect_value = validator.get("expect")
        elif "expected" in validator:
            expect_value = validator.get("expected")
        else:
            raise exceptions.ParamsError("invalid validator: {}".format(validator))

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

    return {
        "check": check_item,
        "expect": expect_value,
        "comparator": comparator
    }


def substitute_variables(content, variables_mapping):
    """ substitute variables in content with variables_mapping

    Args:
        content (str/dict/list/numeric/bool/type): content to be substituted.
        variables_mapping (dict): variables mapping.

    Returns:
        substituted content.

    Examples:
        >>> content = {
                'request': {
                    'url': '/api/users/$uid',
                    'headers': {'token': '$token'}
                }
            }
        >>> variables_mapping = {"$uid": 1000}
        >>> substitute_variables(content, variables_mapping)
            {
                'request': {
                    'url': '/api/users/1000',
                    'headers': {'token': '$token'}
                }
            }

    """
    if isinstance(content, (list, set, tuple)):
        return [
            substitute_variables(item, variables_mapping)
            for item in content
        ]

    if isinstance(content, dict):
        substituted_data = {}
        for key, value in content.items():
            eval_key = substitute_variables(key, variables_mapping)
            eval_value = substitute_variables(value, variables_mapping)
            substituted_data[eval_key] = eval_value

        return substituted_data

    if isinstance(content, basestring):
        # content is in string format here
        for var, value in variables_mapping.items():
            if content == var:
                # content is a variable
                content = value
            else:
                if not isinstance(value, str):
                    value = builtin_str(value)
                content = content.replace(var, value)

    return content

def parse_parameters(parameters, variables_mapping=None, functions_mapping=None):
    """ parse parameters and generate cartesian product.

    Args:
        parameters (list) parameters: parameter name and value in list
            parameter value may be in three types:
                (1) data list, e.g. ["iOS/10.1", "iOS/10.2", "iOS/10.3"]
                (2) call built-in parameterize function, "${parameterize(account.csv)}"
                (3) call custom function in debugtalk.py, "${gen_app_version()}"

        variables_mapping (dict): variables mapping loaded from testcase config
        functions_mapping (dict): functions mapping loaded from debugtalk.py

    Returns:
        list: cartesian product list

    Examples:
        >>> parameters = [
            {"user_agent": ["iOS/10.1", "iOS/10.2", "iOS/10.3"]},
            {"username-password": "${parameterize(account.csv)}"},
            {"app_version": "${gen_app_version()}"}
        ]
        >>> parse_parameters(parameters)

    """
    variables_mapping = variables_mapping or {}
    functions_mapping = functions_mapping or {}
    parsed_parameters_list = []

    for parameter in parameters:
        parameter_name, parameter_content = list(parameter.items())[0]
        parameter_name_list = parameter_name.split("-")

        if isinstance(parameter_content, list):
            # (1) data list
            # e.g. {"app_version": ["2.8.5", "2.8.6"]}
            #       => [{"app_version": "2.8.5", "app_version": "2.8.6"}]
            # e.g. {"username-password": [["user1", "111111"], ["test2", "222222"]}
            #       => [{"username": "user1", "password": "111111"}, {"username": "user2", "password": "222222"}]
            parameter_content_list = []
            for parameter_item in parameter_content:
                if not isinstance(parameter_item, (list, tuple)):
                    # "2.8.5" => ["2.8.5"]
                    parameter_item = [parameter_item]

                # ["app_version"], ["2.8.5"] => {"app_version": "2.8.5"}
                # ["username", "password"], ["user1", "111111"] => {"username": "user1", "password": "111111"}
                parameter_content_dict = dict(zip(parameter_name_list, parameter_item))

                parameter_content_list.append(parameter_content_dict)
        else:
            # (2) & (3)
            parsed_parameter_content = parse_data(parameter_content, variables_mapping, functions_mapping)
            # e.g. [{'app_version': '2.8.5'}, {'app_version': '2.8.6'}]
            # e.g. [{"username": "user1", "password": "111111"}, {"username": "user2", "password": "222222"}]
            if not isinstance(parsed_parameter_content, list):
                raise exceptions.ParamsError("parameters syntax error!")

            parameter_content_list = [
                # get subset by parameter name
                {key: parameter_item[key] for key in parameter_name_list}
                for parameter_item in parsed_parameter_content
            ]

        parsed_parameters_list.append(parameter_content_list)

    return utils.gen_cartesian_product(*parsed_parameters_list)


###############################################################################
##  parse content with variables and functions mapping
###############################################################################

def get_mapping_variable(variable_name, variables_mapping):
    """ get variable from variables_mapping.

    Args:
        variable_name (str): variable name
        variables_mapping (dict): variables mapping

    Returns:
        mapping variable value.

    Raises:
        exceptions.VariableNotFound: variable is not found.

    """
    try:
        return variables_mapping[variable_name]
    except KeyError:
        raise exceptions.VariableNotFound("{} is not found.".format(variable_name))


def get_mapping_function(function_name, functions_mapping):
    """ get function from functions_mapping,
        if not found, then try to check if builtin function.

    Args:
        variable_name (str): variable name
        variables_mapping (dict): variables mapping

    Returns:
        mapping function object.

    Raises:
        exceptions.FunctionNotFound: function is neither defined in debugtalk.py nor builtin.

    """
    if function_name in functions_mapping:
        return functions_mapping[function_name]

    try:
        # check if HttpRunner builtin functions
        from httprunner import loader
        built_in_functions = loader.load_builtin_functions()
        return built_in_functions[function_name]
    except KeyError:
        pass

    try:
        # check if Python builtin functions
        item_func = eval(function_name)
        if callable(item_func):
            # is builtin function
            return item_func
    except (NameError, TypeError):
        # is not builtin function
        raise exceptions.FunctionNotFound("{} is not found.".format(function_name))


def parse_string_functions(content, variables_mapping, functions_mapping):
    """ parse string content with functions mapping.

    Args:
        content (str): string content to be parsed.
        variables_mapping (dict): variables mapping.
        functions_mapping (dict): functions mapping.

    Returns:
        str: parsed string content.

    Examples:
        >>> content = "abc${add_one(3)}def"
        >>> functions_mapping = {"add_one": lambda x: x + 1}
        >>> parse_string_functions(content, functions_mapping)
            "abc4def"

    """
    functions_list = extract_functions(content)
    for func_content in functions_list:
        function_meta = parse_function(func_content)
        func_name = function_meta["func_name"]

        args = function_meta.get("args", [])
        kwargs = function_meta.get("kwargs", {})
        args = parse_data(args, variables_mapping, functions_mapping)
        kwargs = parse_data(kwargs, variables_mapping, functions_mapping)

        if func_name in ["parameterize", "P"]:
            if len(args) != 1 or kwargs:
                raise exceptions.ParamsError("P() should only pass in one argument!")
            from httprunner import loader
            eval_value = loader.load_csv_file(args[0])
        elif func_name in ["environ", "ENV"]:
            if len(args) != 1 or kwargs:
                raise exceptions.ParamsError("ENV() should only pass in one argument!")
            eval_value = utils.get_os_environ(args[0])
        else:
            func = get_mapping_function(func_name, functions_mapping)
            eval_value = func(*args, **kwargs)

        func_content = "${" + func_content + "}"
        if func_content == content:
            # content is a function, e.g. "${add_one(3)}"
            content = eval_value
        else:
            # content contains one or many functions, e.g. "abc${add_one(3)}def"
            content = content.replace(
                func_content,
                str(eval_value), 1
            )

    return content


def parse_string_variables(content, variables_mapping):
    """ parse string content with variables mapping.

    Args:
        content (str): string content to be parsed.
        variables_mapping (dict): variables mapping.

    Returns:
        str: parsed string content.

    Examples:
        >>> content = "/api/users/$uid"
        >>> variables_mapping = {"$uid": 1000}
        >>> parse_string_variables(content, variables_mapping)
            "/api/users/1000"

    """
    variables_list = extract_variables(content)
    for variable_name in variables_list:
        variable_value = get_mapping_variable(variable_name, variables_mapping)

        # TODO: replace variable label from $var to {{var}}
        if "${}".format(variable_name) == content:
            # content is a variable
            content = variable_value
        else:
            # content contains one or several variables
            if not isinstance(variable_value, str):
                variable_value = builtin_str(variable_value)

            content = content.replace(
                "${}".format(variable_name),
                variable_value, 1
            )

    return content


def parse_data(content, variables_mapping=None, functions_mapping=None):
    """ parse content with variables mapping

    Args:
        content (str/dict/list/numeric/bool/type): content to be parsed
        variables_mapping (dict): variables mapping.
        functions_mapping (dict): functions mapping.

    Returns:
        parsed content.

    Examples:
        >>> content = {
                'request': {
                    'url': '/api/users/$uid',
                    'headers': {'token': '$token'}
                }
            }
        >>> variables_mapping = {"uid": 1000, "token": "abcdef"}
        >>> parse_data(content, variables_mapping)
            {
                'request': {
                    'url': '/api/users/1000',
                    'headers': {'token': 'abcdef'}
                }
            }

    """
    # TODO: refactor type check
    if content is None or isinstance(content, (numeric_types, bool, type)):
        return content

    if isinstance(content, (list, set, tuple)):
        return [
            parse_data(item, variables_mapping, functions_mapping)
            for item in content
        ]

    if isinstance(content, dict):
        parsed_content = {}
        for key, value in content.items():
            parsed_key = parse_data(key, variables_mapping, functions_mapping)
            parsed_value = parse_data(value, variables_mapping, functions_mapping)
            parsed_content[parsed_key] = parsed_value

        return parsed_content

    if isinstance(content, basestring):
        # content is in string format here
        variables_mapping = utils.ensure_mapping_format(variables_mapping or {})
        functions_mapping = functions_mapping or {}
        content = content.strip()

        # replace functions with evaluated value
        # Notice: _eval_content_functions must be called before _eval_content_variables
        content = parse_string_functions(content, variables_mapping, functions_mapping)

        # replace variables with binding value
        content = parse_string_variables(content, variables_mapping)

    return content


def _extend_with_api(test_dict, api_def_dict):
    """ extend test with api definition, test will merge and override api definition.

    Args:
        test_dict (dict): test block
        api_def_dict (dict): api definition

    Returns:
        dict: extended test dict.

    Examples:
        >>> api_def_dict = {
                "name": "get token 1",
                "request": {...},
                "validate": [{'eq': ['status_code', 200]}]
            }
        >>> test_dict = {
                "name": "get token 2",
                "extract": {"token": "content.token"},
                "validate": [{'eq': ['status_code', 201]}, {'len_eq': ['content.token', 16]}]
            }
        >>> _extend_with_api(test_dict, api_def_dict)
            {
                "name": "get token 2",
                "request": {...},
                "extract": {"token": "content.token"},
                "validate": [{'eq': ['status_code', 201]}, {'len_eq': ['content.token', 16]}]
            }

    """
    # override name
    api_def_name = api_def_dict.pop("name", "")
    test_dict["name"] = test_dict.get("name") or api_def_name

    # override variables
    def_variables = api_def_dict.pop("variables", [])
    test_dict["variables"] = utils.extend_variables(
        def_variables,
        test_dict.get("variables", {})
    )

    # merge & override validators TODO: relocate
    def_raw_validators = api_def_dict.pop("validate", [])
    ref_raw_validators = test_dict.get("validate", [])
    def_validators = [
        parse_validator(validator)
        for validator in def_raw_validators
    ]
    ref_validators = [
        parse_validator(validator)
        for validator in ref_raw_validators
    ]
    test_dict["validate"] = utils.extend_validators(
        def_validators,
        ref_validators
    )

    # merge & override extractors
    def_extrators = api_def_dict.pop("extract", {})
    test_dict["extract"] = utils.extend_variables(
        def_extrators,
        test_dict.get("extract", {})
    )

    # TODO: merge & override request
    test_dict["request"] = api_def_dict.pop("request", {})

    # base_url & verify: priority api_def_dict > test_dict
    if api_def_dict.get("base_url"):
        test_dict["base_url"] = api_def_dict["base_url"]

    if "verify" in api_def_dict:
        test_dict["request"]["verify"] = api_def_dict["verify"]

    # merge & override setup_hooks
    def_setup_hooks = api_def_dict.pop("setup_hooks", [])
    ref_setup_hooks = test_dict.get("setup_hooks", [])
    extended_setup_hooks = list(set(def_setup_hooks + ref_setup_hooks))
    if extended_setup_hooks:
        test_dict["setup_hooks"] = extended_setup_hooks
    # merge & override teardown_hooks
    def_teardown_hooks = api_def_dict.pop("teardown_hooks", [])
    ref_teardown_hooks = test_dict.get("teardown_hooks", [])
    extended_teardown_hooks = list(set(def_teardown_hooks + ref_teardown_hooks))
    if extended_teardown_hooks:
        test_dict["teardown_hooks"] = extended_teardown_hooks

    # TODO: extend with other api definition items, e.g. times
    test_dict.update(api_def_dict)

    return test_dict


def _extend_with_testcase(test_dict, testcase_def_dict):
    """ extend test with testcase definition
        test will merge and override testcase config definition.

    Args:
        test_dict (dict): test block
        testcase_def_dict (dict): testcase definition

    Returns:
        dict: extended test dict.

    """
    # override testcase config variables
    testcase_def_dict["config"].setdefault("variables", {})
    testcase_def_variables = utils.ensure_mapping_format(testcase_def_dict["config"].get("variables", {}))
    testcase_def_variables.update(test_dict.pop("variables", {}))
    testcase_def_dict["config"]["variables"] = testcase_def_variables

    # override base_url, verify
    # priority: testcase config > testsuite tests
    test_base_url = test_dict.pop("base_url", "")
    if not testcase_def_dict["config"].get("base_url"):
        testcase_def_dict["config"]["base_url"] = test_base_url

    test_verify = test_dict.pop("verify", True)
    testcase_def_dict["config"].setdefault("verify", test_verify)

    # override testcase config name, output, etc.
    testcase_def_dict["config"].update(test_dict)

    test_dict.clear()
    test_dict.update(testcase_def_dict)


def __parse_config(config, project_mapping):
    """ parse testcase config, include variables and name.
    """
    # get config variables
    raw_config_variables = config.pop("variables", {})
    raw_config_variables_mapping = utils.ensure_mapping_format(raw_config_variables)
    override_variables = utils.deepcopy_dict(project_mapping.get("variables", {}))
    functions = project_mapping.get("functions", {})

    # override testcase config variables with passed in variables
    for key, value in raw_config_variables_mapping.items():

        if key in override_variables:
            # passed in
            continue
        else:
            # config variables
            try:
                parsed_value = parse_data(
                    value,
                    override_variables,
                    functions
                )
            except exceptions.VariableNotFound:
                pass
            override_variables[key] = parsed_value

    if override_variables:
        config["variables"] = override_variables

    # parse config name
    config["name"] = parse_data(
        config.get("name", ""),
        override_variables,
        functions
    )

    # parse config base_url
    if "base_url" in config:
        config["base_url"] = parse_data(
            config["base_url"],
            override_variables,
            functions
        )


def __parse_tests(tests, config, project_mapping):
    """ override tests with testcase config variables, base_url and verify.
        test maybe nested testcase.

        variables priority:
        testsuite config > testsuite test > testcase config > testcase test > api

        base_url/verify priority:
        testcase test > testcase config > testsuite test > testsuite config > api

    Args:
        tests (list):
        config (dict):

    Returns:
        list: overrided tests

    """
    config_variables = config.pop("variables", {})
    config_base_url = config.pop("base_url", "")
    config_verify = config.pop("verify", True)
    functions = project_mapping.get("functions", {})

    for test_dict in tests:

        # base_url & verify: priority test_dict > config
        if (not test_dict.get("base_url")) and config_base_url:
            test_dict["base_url"] = config_base_url

        test_dict.setdefault("verify", config_verify)

        if "testcase_def" in test_dict:
            # test_dict is nested testcase

            # 1, testsuite config => testsuite tests
            # override test_dict variables
            test_dict["variables"] = utils.extend_variables(
                test_dict.pop("variables", {}),
                config_variables
            )

            # parse test_dict name
            try:
                test_dict["name"] = parse_data(
                    test_dict.pop("name", ""),
                    test_dict["variables"],
                    functions
                )
            except exceptions.VariableNotFound:
                pass

            # 2, testsuite test_dict => testcase config
            testcase_def = test_dict.pop("testcase_def")
            _extend_with_testcase(test_dict, testcase_def)

            # 3, testcase config => testcase test_dict
            _parse_testcase(test_dict, project_mapping)

        else:
            # 1, config => tests
            # override test_dict variables
            test_dict["variables"] = utils.extend_variables(
                test_dict.pop("variables", {}),
                config_variables
            )

            # parse test_dict name
            try:
                test_dict["name"] = parse_data(
                    test_dict.pop("name", ""),
                    test_dict["variables"],
                    functions
                )
            except exceptions.VariableNotFound:
                pass

            if "api_def" in test_dict:
                # test_dict has API reference
                # 2, test_dict => api
                api_def_dict = test_dict.pop("api_def")
                _extend_with_api(test_dict, api_def_dict)

            if test_dict.get("base_url"):
                base_url = parse_data(
                    test_dict.pop("base_url"),
                    test_dict["variables"],
                    functions
                )
                try:
                    request_url = parse_data(
                        test_dict["request"]["url"],
                        test_dict["variables"],
                        functions
                    )
                except exceptions.VariableNotFound:
                    """ variable in current url maybe extracted from former api
                        "tests": [
                            {
                                "request": {},
                                "extract": {
                                    "host1": content.host1
                                }
                            },
                            {
                                "request": {
                                    "url": "$host1/path1"
                                }
                            }
                        ]
                    """
                    request_url = test_dict["request"]["url"]
                finally:
                    test_dict["request"]["url"] = utils.build_url(
                        base_url,
                        request_url
                    )


def _parse_testcase(testcase, project_mapping):
    __parse_config(testcase["config"], project_mapping)
    __parse_tests(testcase["tests"], testcase["config"], project_mapping)


def parse_tests(tests_mapping):
    """ parse testcases configs, including variables/name/request.

    Args:
        tests_mapping (dict): project info and testcases list.

            {
                "project_mapping": {
                    "PWD": "XXXXX",
                    "functions": {},
                    "variables": {},                        # optional, priority 1
                    "env": {}
                },
                "testcases": [
                    {   # testcase data structure
                        "config": {
                            "name": "desc1",
                            "path": "testcase1_path",
                            "variables": [],                # optional, priority 2
                        },
                        "tests": [
                            # test data structure
                            {
                                'name': 'test step desc1',
                                'variables': [],            # optional, priority 3
                                'extract': [],
                                'validate': [],
                                'api_def': {
                                    "variables": {}         # optional, priority 4
                                    'request': {},
                                }
                            },
                            test_dict_2   # another test dict
                        ]
                    },
                    testcase_dict_2     # another testcase dict
                ]
            }

    """
    project_mapping = tests_mapping.get("project_mapping", {})

    for testcase in tests_mapping["testcases"]:
        testcase.setdefault("config", {})
        _parse_testcase(testcase, project_mapping)
