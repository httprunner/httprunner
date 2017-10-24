#encoding: utf-8
import os
import re
from setuptools import setup, find_packages

# parse version from ate/__init__.py
with open(os.path.join(os.path.dirname(__file__), 'ate', '__init__.py')) as f:
    version = re.compile(r"__version__\s+=\s+'(.*)'", re.I).match(f.read()).group(1)

setup(
    name='ApiTestEngine',
    version=version,
    description='API test engine.',
    long_description="Best practice of API test, including automation test and performance test.",
    author='Leo Lee',
    author_email='mail@debugtalk.com',
    url='https://github.com/debugtalk/ApiTestEngine',
    license='MIT',
    packages=find_packages(exclude=['test.*', 'test']),
    package_data={
        'ate': ['locustfile_template'],
    },
    keywords='api test',
    install_requires=[
        "requests[security]",
        "flask",
        "PyYAML",
        "coveralls",
        "coverage",
        "PyUnitReport"
    ],
    extras_require={
        'locustio': [
            "locustio"
        ]
    },
    dependency_links=[
        "git+https://github.com/debugtalk/PyUnitReport.git#egg=PyUnitReport-0",
        "git+https://github.com/locustio/locust.git#egg=locust-0"
    ],
    classifiers=[
        "Development Status :: 3 - Alpha",
        'Programming Language :: Python :: 2.7',
        'Programming Language :: Python :: 3.4',
        'Programming Language :: Python :: 3.5',
        'Programming Language :: Python :: 3.6'
    ],
    entry_points={
        'console_scripts': [
            'ate=ate.cli:main_ate',
            'locusts=ate.cli:main_locust'
        ]
    }
)
