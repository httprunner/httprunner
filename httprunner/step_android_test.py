from httprunner import HttpRunner, Config, Step, RunAndroidUI


class TestCaseAndroidDemo(HttpRunner):

    config = (
        Config("demo for android UI test")
        .variables(
            **{
                "foo1": "config_bar1",
                "foo2": "config_bar2",
                "expect_foo1": "config_bar1",
                "expect_foo2": "config_bar2",
            }
        )
        .android()
        .serial("xxx")
        .package_name("xxx")
        .install_apk("xxx")
    )

    teststeps = [
        # Step(
        #     RunAndroidUI("start app").control().start_app("com.ss.android.ugc.aweme")
        # ),
        Step(
            RunAndroidUI("back home").ui().press_home()
        ),
        Step(
            RunAndroidUI("back home").control().start_app()
        ),
        Step(
            RunAndroidUI("swipe up").ui().swipe_up()
        ),
        Step(
            RunAndroidUI("swipe up").ui().swipe_up()
        ),
    ]


if __name__ == "__main__":
    TestCaseAndroidDemo().test_start()
