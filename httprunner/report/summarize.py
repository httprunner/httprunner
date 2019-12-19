import platform

from httprunner import __version__


def get_platform():
    return {
        "httprunner_version": __version__,
        "python_version": "{} {}".format(
            platform.python_implementation(),
            platform.python_version()
        ),
        "platform": platform.platform()
    }


def aggregate_stat(origin_stat, new_stat):
    """ aggregate new_stat to origin_stat.

    Args:
        origin_stat (dict): origin stat dict, will be updated with new_stat dict.
        new_stat (dict): new stat dict.

    """
    for key in new_stat:
        if key not in origin_stat:
            origin_stat[key] = new_stat[key]
        elif key == "start_at":
            # start datetime
            origin_stat["start_at"] = min(origin_stat["start_at"], new_stat["start_at"])
        elif key == "duration":
            # duration = max_end_time - min_start_time
            max_end_time = max(origin_stat["start_at"] + origin_stat["duration"],
                               new_stat["start_at"] + new_stat["duration"])
            min_start_time = min(origin_stat["start_at"], new_stat["start_at"])
            origin_stat["duration"] = max_end_time - min_start_time
        else:
            origin_stat[key] += new_stat[key]


def get_summary(result):
    """ get summary from test result

    Args:
        result (instance): HtmlTestResult() instance

    Returns:
        dict: summary extracted from result.

            {
                "success": True,
                "stat": {},
                "time": {},
                "records": []
            }

    """
    summary = {
        "success": result.wasSuccessful(),
        "stat": {
            'total': result.testsRun,
            'failures': len(result.failures),
            'errors': len(result.errors),
            'skipped': len(result.skipped),
            'expectedFailures': len(result.expectedFailures),
            'unexpectedSuccesses': len(result.unexpectedSuccesses)
        }
    }
    summary["stat"]["successes"] = summary["stat"]["total"] \
                                   - summary["stat"]["failures"] \
                                   - summary["stat"]["errors"] \
                                   - summary["stat"]["skipped"] \
                                   - summary["stat"]["expectedFailures"] \
                                   - summary["stat"]["unexpectedSuccesses"]

    summary["time"] = {
        'start_at': result.start_at,
        'duration': result.duration
    }
    summary["records"] = result.records

    return summary
