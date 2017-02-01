# Copyright (c) 2016 Intel Corporation
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.
# -*- coding: utf-8 -*-
"""
Module to launch cppr_daemon on all specified nodes via orterun
"""
import os
import subprocess
import tempfile
import re


class CpprDaemonRunner:
    """
    Functionality to launch + kill cppr_daemon on all specified nodes
    """
    info = None

    def __init__(self, info=None):
        self.info = info

    def launch_process(self):
        """
        Launch cppr_daemon on all nodes specified in host_list
        """
        print("TestRunner: start cppr_daemon process.")

        w_path = tempfile.mkdtemp()
        os.environ["CNSS_PREFIX"] = os.path.join(w_path, "local")

        self.info.set_config('setKeyFromConfig', "CNSS_PREFIX",
                             os.path.join(w_path, "local"))
        g_path = os.path.join(os.path.expanduser("~/"), "tmp")
        self.info.set_config('setKeyFromConfig', "GLOBAL_PATH",
                             os.path.join(g_path, "global"))
        if not os.path.exists(w_path):
            os.makedirs(w_path)
        if not os.path.exists(g_path):
            os.makedirs(g_path)

        hosts = ","
        hostlist = hosts.join(self.info.get_config('host_list'))
        daemon_exe = \
            os.path.join(
                self.info.get_info("PREFIX"),
                "TESTING",
                "scripts",
                "run_cppr.sh")

        cmdarg = [daemon_exe, "-w", w_path, "-g", g_path, "-H",
                  hostlist, "-l", "4", "start"]
        log_path = self.info.get_config('log_base_path')
        fileout = os.path.join(log_path, "cppr-start.out")
        filerr = os.path.join(log_path, "cppr-start.err")
        with open(fileout, mode='w+b', buffering=0) as outfile, \
            open(filerr, mode='w+b', buffering=0) as errfile:
            rc = subprocess.call(cmdarg, stdout=outfile, stderr=errfile)

        return rc

    def stop_process(self):
        """
        Kill cppr_daemon on all nodes
        """
        print("TestRunner: stopping cppr_daemon process.")
        w_path = os.getenv('CNSS_PREFIX', "")
        re.sub('/local', '', w_path)
        if not w_path:
            print("CNSS_PREFIX is not set correctly.")
            raise ValueError
        g_path = os.path.join(os.path.expanduser("~/"), "tmp")
        hosts = ","
        hostlist = hosts.join(self.info.get_config('host_list'))
        daemon_exe = \
            os.path.join(
                self.info.get_info("PREFIX"),
                "TESTING",
                "scripts",
                "run_cppr.sh")

        cmdarg = [daemon_exe, "-w", w_path, "-g", g_path, "-H",
                  hostlist, "stop"]
        log_path = self.info.get_config('log_base_path')
        fileout = os.path.join(log_path, "cppr-stop.out")
        filerr = os.path.join(log_path, "cppr-stop.err")
        with open(fileout, mode='w+b', buffering=0) as outfile, \
            open(filerr, mode='w+b', buffering=0) as errfile:
            rc = subprocess.check_call(cmdarg, stdout=outfile, stderr=errfile)

        return rc
