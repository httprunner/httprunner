# encoding: utf-8

import copy
import io
import itertools
import json
import os.path
import string
from datetime import datetime

from httprunner import exceptions, logger
from httprunner.compat import OrderedDict, basestring, is_py2


def remove_prefix(text, prefix):
    """ remove prefix from text
    """
    if text.startswith(prefix):
        return text[len(prefix):]
    return text


def set_os_environ(variables_mapping):
    """ set variables mapping to os.environ
    """
    for variable in variables_mapping:
        os.environ[variable] = variables_mapping[variable]
        logger.log_debug("Loaded variable: {}".format(variable))


def query_json(json_content, query, delimiter='.'):
    """ Do an xpath-like query with json_content.
    @param (dict/list/string) json_content
        json_content = {
            "ids": [1, 2, 3, 4],
            "person": {
                "name": {
                    "first_name": "Leo",
                    "last_name": "Lee",
                },
                "age": 29,
                "cities": ["Guangzhou", "Shenzhen"]
            }
        }
    @param (str) query
        "person.name.first_name"  =>  "Leo"
        "person.name.first_name.0"  =>  "L"
        "person.cities.0"         =>  "Guangzhou"
    @return queried result
    """
    raise_flag = False
    response_body = u"response body: {}\n".format(json_content)
    try:
        for key in query.split(delimiter):
            if isinstance(json_content, (list, basestring)):
                json_content = json_content[int(key)]
            elif isinstance(json_content, dict):
                json_content = json_content[key]
            else:
                logger.log_error(
                    "invalid type value: {}({})".format(json_content, type(json_content)))
                raise_flag = True
    except (KeyError, ValueError, IndexError):
        raise_flag = True

    if raise_flag:
        err_msg = u"Failed to extract! => {}\n".format(query)
        err_msg += response_body
        logger.log_error(err_msg)
        raise exceptions.ExtractFailure(err_msg)

    return json_content


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
    elif comparator in ["len_ge", "count_ge", "length_greater_than_or_equals", \
        "count_greater_than_or_equals"]:
        return "length_greater_than_or_equals"
    elif comparator in ["len_lt", "count_lt", "length_less_than", "count_less_than"]:
        return "length_less_than"
    elif comparator in ["len_le", "count_le", "length_less_than_or_equals", \
        "count_less_than_or_equals"]:
        return "length_less_than_or_equals"
    else:
        return comparator

def deep_update_dict(origin_dict, override_dict):
    """ update origin dict with override dict recursively
    e.g. origin_dict = {'a': 1, 'b': {'c': 2, 'd': 4}}
         override_dict = {'b': {'c': 3}}
    return: {'a': 1, 'b': {'c': 3, 'd': 4}}
    """
    if not override_dict:
        return origin_dict

    for key, val in override_dict.items():
        if isinstance(val, dict):
            tmp = deep_update_dict(origin_dict.get(key, {}), val)
            origin_dict[key] = tmp
        elif val is None:
            # fix #64: when headers in test is None, it should inherit from config
            continue
        else:
            origin_dict[key] = override_dict[key]

    return origin_dict

def lower_dict_keys(origin_dict):
    """ convert keys in dict to lower case
    e.g.
        Name => name, Request => request
        URL => url, METHOD => method, Headers => headers, Data => data
    """
    if not origin_dict or not isinstance(origin_dict, dict):
        return origin_dict

    return {
        key.lower(): value
        for key, value in origin_dict.items()
    }

def lower_config_dict_key(config_dict):
    """ convert key in config dict to lower case, convertion will occur in three places:
        1, all keys in config dict;
        2, all keys in config["request"]
        3, all keys in config["request"]["headers"]
    """
    # convert keys in config dict
    config_dict = lower_dict_keys(config_dict)

    if "request" in config_dict:
        # convert keys in config["request"]
        config_dict["request"] = lower_dict_keys(config_dict["request"])

        # convert keys in config["request"]["headers"]
        if "headers" in config_dict["request"]:
            config_dict["request"]["headers"] = lower_dict_keys(config_dict["request"]["headers"])

    return config_dict

def convert_mappinglist_to_orderdict(mapping_list):
    """ convert mapping list to ordered dict

    Args:
        mapping_list (list):
            [
                {"a": 1},
                {"b": 2}
            ]

    Returns:
        OrderedDict: converted mapping in OrderedDict
            OrderDict(
                {
                    "a": 1,
                    "b": 2
                }
            )

    """
    ordered_dict = OrderedDict()
    for map_dict in mapping_list:
        ordered_dict.update(map_dict)

    return ordered_dict


def update_ordered_dict(ordered_dict, override_mapping):
    """ override ordered_dict with new mapping.

    Args:
        ordered_dict (OrderDict): original ordered dict
        override_mapping (dict): new variables mapping

    Returns:
        OrderDict: new overrided variables mapping.

    Examples:
        >>> ordered_dict = OrderDict({"a": 1, "b": 2})
        >>> override_mapping = {"a": 3, "c": 4}
        >>> update_ordered_dict(ordered_dict, override_mapping)
            OrderDict({"a": 3, "b": 2, "c": 4})

    """
    new_ordered_dict = copy.copy(ordered_dict)
    for var, value in override_mapping.items():
        new_ordered_dict.update({var: value})

    return new_ordered_dict


def override_mapping_list(variables, new_mapping):
    """ override variables with new mapping.

    Args:
        variables (list): variables list
            [
                {"var_a": 1},
                {"var_b": "world"}
            ]
        new_mapping (dict): overrided variables mapping
            {
                "var_a": "hello"
            }

    Returns:
        OrderedDict: overrided variables mapping.

    Examples:
        >>> variables = [
                {"var_a": 1},
                {"var_b": "world"}
            ]
        >>> new_mapping = {
                "var_a": "hello"
            }
        >>> override_mapping_list(variables, new_mapping)
            OrderedDict(
                {
                    "var_a": "hello",
                    "var_b": "world"
                }
            )

    """
    if isinstance(variables, list):
        variables_ordered_dict = convert_mappinglist_to_orderdict(variables)
    elif isinstance(variables, (OrderedDict, dict)):
        variables_ordered_dict = variables
    else:
        raise exceptions.ParamsError("variables error!")

    return update_ordered_dict(
        variables_ordered_dict,
        new_mapping
    )


def add_teststep(test_runner, teststep_dict):
    """ add teststep to testcase.
    """
    def test(self):
        try:
            test_runner.run_test(teststep_dict)
        except exceptions.MyBaseFailure as ex:
            self.fail(str(ex))
        finally:
            if hasattr(test_runner.http_client_session, "meta_data"):
                self.meta_data = test_runner.http_client_session.meta_data
                self.meta_data["validators"] = test_runner.evaluated_validators
                test_runner.http_client_session.init_meta_data()

    test.__doc__ = teststep_dict["name"]
    return test


def get_testcase_io(testcase):
    """ get testcase input(variables) and output.

    Args:
        testcase (unittest.suite.TestSuite): corresponding to one YAML/JSON file, it has been set two attributes:
            config: parsed config block
            runner: initialized runner.Runner() with config

    Returns:
        dict: input(variables) and output mapping.

    """
    runner = testcase.runner
    variables = testcase.config.get("variables", [])
    output_list = testcase.config.get("output", [])

    return {
        "in": dict(variables),
        "out": runner.extract_output(output_list)
    }


def print_io(in_out):
    """ print input(variables) and output.

    Args:
        in_out (dict): input(variables) and output mapping.

    Examples:
        >>> in_out = {
                "in": {
                    "var_a": "hello",
                    "var_b": "world"
                },
                "out": {
                    "status_code": 500
                }
            }
        >>> print_io(in_out)
        ================== Variables & Output ==================
        Type   | Variable         :  Value
        ------ | ---------------- :  ---------------------------
        Var    | var_a            :  hello
        Var    | var_b            :  world

        Out    | status_code      :  500
        --------------------------------------------------------

    """
    content_format = "{:<6} | {:<16} :  {:<}\n"
    content = "\n================== Variables & Output ==================\n"
    content += content_format.format("Type", "Variable", "Value")
    content += content_format.format("-" * 6, "-" * 16, "-" * 27)

    def prepare_content(var_type, in_out):
        content = ""
        for variable, value in in_out.items():
            if isinstance(value, tuple):
                continue
            elif isinstance(value, dict):
                value = json.dumps(value)

            if is_py2:
                if isinstance(variable, unicode):
                    variable = variable.encode("utf-8")
                if isinstance(value, unicode):
                    value = value.encode("utf-8")

            content += content_format.format(var_type, variable, value)

        return content

    _in = in_out["in"]
    _out = in_out["out"]

    content += prepare_content("Var", _in)
    content += "\n"
    content += prepare_content("Out", _out)
    content += "-" * 56 + "\n"

    logger.log_debug(content)

def create_scaffold(project_path):
    if os.path.isdir(project_path):
        folder_name = os.path.basename(project_path)
        logger.log_warning(u"Folder {} exists, please specify a new folder name.".format(folder_name))
        return

    logger.color_print("Start to create new project: {}\n".format(project_path), "GREEN")

    def create_path(path, ptype):
        if ptype == "folder":
            os.makedirs(path)
        elif ptype == "file":
            open(path, 'w').close()

        return "created {}: {}\n".format(ptype, path)

    path_list = [
        (project_path, "folder"),
        (os.path.join(project_path, "api"), "folder"),
        (os.path.join(project_path, "testcases"), "folder"),
        (os.path.join(project_path, "testsuites"), "folder"),
        (os.path.join(project_path, "reports"), "folder"),
        (os.path.join(project_path, "debugtalk.py"), "file"),
        (os.path.join(project_path, ".env"), "file")
    ]

    msg = ""
    for p in path_list:
        msg += create_path(p[0], p[1])

    logger.color_print(msg, "BLUE")


def gen_cartesian_product(*args):
    """ generate cartesian product for lists
    @param
        (list) args
            [{"a": 1}, {"a": 2}],
            [
                {"x": 111, "y": 112},
                {"x": 121, "y": 122}
            ]
    @return
        cartesian product in list
        [
            {'a': 1, 'x': 111, 'y': 112},
            {'a': 1, 'x': 121, 'y': 122},
            {'a': 2, 'x': 111, 'y': 112},
            {'a': 2, 'x': 121, 'y': 122}
        ]
    """
    if not args:
        return []
    elif len(args) == 1:
        return args[0]

    product_list = []
    for product_item_tuple in itertools.product(*args):
        product_item_dict = {}
        for item in product_item_tuple:
            product_item_dict.update(item)

        product_list.append(product_item_dict)

    return product_list


def validate_json_file(file_list):
    """ validate JSON testset format
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


def prettify_json_file(file_list):
    """ prettify JSON testset format
    """
    for json_file in set(file_list):
        if not json_file.endswith(".json"):
            logger.log_warning("Only JSON file format can be prettified, skip: {}".format(json_file))
            continue

        logger.color_print("Start to prettify JSON file: {}".format(json_file), "GREEN")

        dir_path = os.path.dirname(json_file)
        file_name, file_suffix = os.path.splitext(os.path.basename(json_file))
        outfile = os.path.join(dir_path, "{}.pretty.json".format(file_name))

        with io.open(json_file, 'r', encoding='utf-8') as stream:
            try:
                obj = json.load(stream)
            except ValueError as e:
                raise SystemExit(e)

        with io.open(outfile, 'w', encoding='utf-8') as out:
            json.dump(obj, out, indent=4, separators=(',', ': '))
            out.write('\n')

        print("success: {}".format(outfile))


def get_python2_retire_msg():
    retire_day = datetime(2020, 1, 1)
    today = datetime.now()
    left_days = (retire_day - today).days

    if left_days > 0:
        retire_msg = "Python 2 will retire in {} days, why not move to Python 3?".format(left_days)
    else:
        retire_msg = "Python 2 has been retired, you should move to Python 3."

    return retire_msg
