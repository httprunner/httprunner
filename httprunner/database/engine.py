# -*- coding: utf-8 -*-
import datetime
import json

from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker


class DBEngine(object):
    def __init__(self, db_uri):
        """
        db_uri = f'mysql+pymysql://{username}:{password}@{host}:{port}/{database}?charset=utf8mb4'

        """
        engine = create_engine(db_uri)
        self.session = sessionmaker(bind=engine)()

    @staticmethod
    def value_decode(row: dict):
        """
        Try to decode value of table
        datetime.datetime-->string
        datetime.date-->string
        json str-->dict
        :param row:
        :return:
        """
        for k, v in row.items():
            if isinstance(v, datetime.datetime):
                row[k] = v.strftime("%Y-%m-%d %H:%M:%S")
            elif isinstance(v, datetime.date):
                row[k] = v.strftime("%Y-%m-%d")
            elif isinstance(v, str):
                try:
                    row[k] = json.loads(v)
                except ValueError:
                    pass

    def _fetch(self, query, size=-1, commit=True):
        result = self.session.execute(query)
        self.session.commit() if commit else 0
        if query.upper()[:6] == "SELECT":
            if size < 0:
                al = result.fetchall()
                al = [dict(el) for el in al]
                return al or None
            elif size == 1:
                on = dict(result.fetchone())
                self.value_decode(on)
                return on or None
            else:
                mny = result.fetchmany(size)
                mny = [dict(el) for el in mny]
                return mny or None
        elif query.upper()[:6] in ("UPDATE", "DELETE", "INSERT"):
            return {"rowcount": result.rowcount}

    def fetchone(self, query, commit=True):
        return self._fetch(query, size=1, commit=commit)

    def fetchmany(self, query, size, commit=True):
        return self._fetch(query=query, size=size, commit=commit)

    def fetchall(self, query, commit=True):
        return self._fetch(query=query, size=-1, commit=commit)

    def insert(self, query, commit=True):
        return self._fetch(query=query, commit=commit)

    def delete(self, query, commit=True):
        return self._fetch(query=query, commit=commit)

    def update(self, query, commit=True):
        return self._fetch(query=query, commit=commit)


if __name__ == "__main__":
    db = DBEngine(f"mysql+pymysql://xxxxx:xxxxx@10.0.0.1:3306/dbname?charset=utf8mb4")
