#!/usr/bin/env python
# -*- coding: utf-8 -*-

import pytest
import re
import regex

sip_re = re.compile('^["]{0,1}([^"]*)["]{0,1}[ ]*<(sip|tel|sips):(([^@]*)@){0,1}([^>^:]*|\\[[a-fA-F0-9:]*\\]):{0,1}([0-9]*){0,1}>(;.*){0,1}$')
sip_regex = regex.compile('^["]{0,1}([^"]*)["]{0,1}[ ]*<(sip|tel|sips):(([^@]*)@){0,1}([^>^:]*|\\[[a-fA-F0-9:]*\\]):{0,1}([0-9]*){0,1}>(;.*){0,1}$')
test_strings = [
    "\"display_name\"<sip:0312341234@10.0.0.1:5060>;user=phone;hogehoge",
	"<sip:0312341234@10.0.0.1>",
	"\"display_name\"<sip:0312341234@10.0.0.1>",
	"<sip:whois.this>;user=phone",
	"\"0333334444\"<sip:[2001:30:fe::4:123]>;user=phone"
]

def exec_re():
    for s in test_strings:
        sip_re.match(s)
    return True

def exec_regex():
    for s in test_strings:
        sip_regex.match(s)
    return True

def test_re_benchmark(benchmark):
    assert benchmark(exec_re)

def test_regex_benchmark(benchmark):
    assert benchmark(exec_regex)


if __name__ == '__main__':
    pytest.main(['-v', __file__])