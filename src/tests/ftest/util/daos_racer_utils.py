"""
  (C) Copyright 2020-2022 Intel Corporation.

  SPDX-License-Identifier: BSD-2-Clause-Patent
"""
import os

from ClusterShell.NodeSet import NodeSet

from command_utils_base import BasicParameter, FormattedParameter
from exception_utils import CommandFailure, MPILoadError
from command_utils import ExecutableCommand
from general_utils import pcmd, get_log_file
from env_modules import load_mpi


class DaosRacerCommand(ExecutableCommand):
    """Defines a object representing a daos_racer command."""

    def __init__(self, path, host, dmg=None):
        """Create a daos_racer command object.

        Args:
            path (str): path of the daos_racer command
            host (str): host on which to run the daos_racer command
            dmg (DmgCommand): a DmgCommand object used to obtain the
                configuration file and certificate
        """
        super().__init__(
            "/run/daos_racer/*", "daos_racer", path)
        self.host = NodeSet(host)

        # Number of seconds to run
        self.runtime = FormattedParameter("-t {}", 60)
        self.pool_uuid = FormattedParameter("-p {}")
        self.cont_uuid = FormattedParameter("-c {}")

        if dmg:
            self.dmg_config = FormattedParameter("-n {}", dmg.yaml.filename)
            dmg.copy_certificates(get_log_file("daosCA/certs"), self.host)
            dmg.copy_configuration(self.host)

        # Optional timeout for the clush command running the daos_racer command.
        # This should be set greater than the 'runtime' value but less than the
        # avocado test timeout value to allow for proper cleanup.  Using a value
        # of None will result in no timeout being used.
        self.clush_timeout = BasicParameter(None)

    @property
    def pre_command(self):
        """Get any pre-command exports or commands."""
        return self.env.to_export_str()

    def get_str_param_names(self):
        """Get a sorted list of the names of the command attributes.

        Only include FormattedParameter class parameter values when building the
        command string, e.g. 'runtime'.

        Returns:
            list: a list of class attribute names used to define parameters
                for the command.

        """
        return self.get_attribute_names(FormattedParameter)

    def get_environment(self, log_file=None):
        """Get the environment variables to export for the daos_racer command.

        Args:
            log_file (str, optional): when specified overrides the default
                D_LOG_FILE value. Defaults to None.

        Returns:
            EnvironmentVariables: a dictionary of environment variable names and
                values to export prior to running daos_racer

        """
        env = self.env.copy()
        env['D_LOG_FILE'] = get_log_file(log_file or "{}_daos.log".format(self.command))
        env["OMPI_MCA_btl_openib_warn_default_gid_prefix"] = "0"
        env["OMPI_MCA_btl"] = "tcp,self"
        env["OMPI_MCA_oob"] = "tcp"
        env["OMPI_MCA_pml"] = "ob1"
        env["D_LOG_MASK"] = "ERR"

        if not load_mpi("openmpi"):
            raise MPILoadError("openmpi")

        env["LD_LIBRARY_PATH"] = os.environ["LD_LIBRARY_PATH"]

        return env

    def run(self):
        """Run the daos_racer command remotely.

        Raises:
            CommandFailure: if there is an error running the command

        """
        # Run daos_racer on the specified host
        self.log.info(
            "Running %s on %s with %s timeout",
            str(self), self.host,
            "no" if self.clush_timeout.value is None else
            "a {}s".format(self.clush_timeout.value))
        return_codes = pcmd(self.host, str(self), True, self.clush_timeout.value)
        if 0 not in return_codes or len(return_codes) > 1:
            # Kill the daos_racer process if the remote command timed out
            if 255 in return_codes:
                self.log.info(
                    "Stopping timed out daos_racer process on %s", self.host)
                pcmd(self.host, "pkill daos_racer", True)

            raise CommandFailure("Error running '{}'".format(self._command))

        self.log.info("Test passed!")
