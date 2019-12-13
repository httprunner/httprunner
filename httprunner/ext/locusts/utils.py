from httprunner import loader, parser


def prepare_locust_tests(path):
    """ prepare locust testcases

    Args:
        path (str): testcase file path.

    Returns:
        list: locust tests data

            [
                testcase1_dict,
                testcase2_dict
            ]

    """
    tests_mapping = loader.load_cases(path)
    testcases = parser.parse_tests(tests_mapping)

    locust_tests = []

    for testcase in testcases:
        testcase_weight = testcase.get("config", {}).pop("weight", 1)
        for _ in range(testcase_weight):
            locust_tests.append(testcase)

    return locust_tests
