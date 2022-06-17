import sys
from pathlib import Path

from httprunner.database.engine import DBEngine

sys.path.insert(0, str(Path(__file__).parent.parent))

from httprunner import HttpRunner, Config, Step, RunSqlRequest  # noqa:E402


class TestCaseDemoSqlite(HttpRunner):
    config = Config("run sqlite demo")

    teststeps = [
        Step(
            RunSqlRequest("执行一个sqlite demo")
            .fetchmany("select* from student;", 5)
            .extract()
            .with_jmespath("[0].name", "name")
            .validate()
            .assert_equal(
                "[0]",
                {
                    "id": 1,
                    "name": "Jack",
                    "fullname": {"first_name": "Jack", "last_name": "Tomson"},
                },
            )
            .assert_equal("[0].fullname.first_name", "Jack")
        )
    ]

    def test_start(self):
        eg = DBEngine(db_uri="sqlite:///../data/sqlite.db")
        self.with_db_engine(eg)
        super().test_start()
