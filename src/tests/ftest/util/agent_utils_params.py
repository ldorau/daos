#!/usr/bin/python
"""
  (C) Copyright 2020-2022 Intel Corporation.

  SPDX-License-Identifier: BSD-2-Clause-Patent
"""
import os

from command_utils_base import \
    BasicParameter, LogParameter, YamlParameters, TransportCredentials


class DaosAgentTransportCredentials(TransportCredentials):
    """Transport credentials listing certificates for secure communication."""

    def __init__(self, log_dir="/tmp"):
        """Initialize a TransportConfig object."""
        super().__init__(
            "/run/agent_config/transport_config/*", "transport_config", log_dir)

        # Additional daos_agent transport credential parameters:
        #   - server_name: <str>, e.g. "daos_server"
        #       Name of server according to its certificate [daos_agent only]
        #
        self.server_name = BasicParameter(None, None)
        self.cert = LogParameter(log_dir, None, "agent.crt")
        self.key = LogParameter(log_dir, None, "agent.key")


class DaosAgentYamlParameters(YamlParameters):
    """Defines the daos_agent configuration yaml parameters."""

    # TODO model after server_utils_params
    class FabricIfaceDeviceYamlParameters(YamlParameters):
        """Defines the fabric_iface device yaml parameters."""

        def __init__(self, iface_index=None, device_index=None):
            """Initialize a FabricIfaceDeviceYamlParameters object.

            Args:
                iface_index (int, optional): index number for the namespace path used
                    when specifying multiple fabric_ifaces. Defaults to None.
                device_index (int, optional): index number for the namespace path used
                    when specifying multiple devices. Defaults to None.
            """
            iface_str = "/{}".format(iface_index) if iface_index is not None else ""
            dev_str = "/{}".format(device_index) if device_index is not None else ""
            namespace = "/run/agent_config/fabric_ifaces{}/devices{}/*".format(iface_str, dev_str)
            super().__init__(namespace)

            # Parameters
            #   iface:              TODO
            #   domain:             TODO
            self.iface = BasicParameter(None)
            self.domain = BasicParameter(None)

    class FabricIfaceYamlParameters(YamlParameters):
        """Defines the fabric_iface yaml parameters."""

        def __init__(self, iface_index=None, device_index=None):
            """Initialize a FabricIfaceYamlParameters object.

            Args:
                iface_index (int, optional): index number for the namespace path used
                    when specifying multiple fabric_ifaces. Defaults to None.
                device_index (int, optional): index number for the namespace path used
                    when specifying multiple devices. Defaults to None.
            """
            iface_str = "/{}".format(iface_index) if iface_index is not None else ""
            dev_str = "/{}".format(device_index) if device_index is not None else ""
            namespace = "/run/agent_config/fabric_ifaces{}/devices{}/*".format(iface_str, dev_str)
            super().__init__(namespace)

            # Parameters
            #   iface:              TODO
            #   domain:             TODO
            self.iface = BasicParameter(None)
            self.domain = BasicParameter(None)

    def __init__(self, filename, common_yaml):
        """Initialize an DaosAgentYamlParameters object.

        Args:
            filename (str): yaml configuration file name
            common_yaml (YamlParameters): [description]
        """
        super().__init__("/run/agent_config/*", filename, None, common_yaml)

        # All log files should be placed in the same directory on each host to
        # enable easy log file archiving by launch.py
        log_dir = os.environ.get("DAOS_TEST_LOG_DIR", "/tmp")

        # daos_agent parameters:
        #   - runtime_dir: <str>, e.g. /var/run/daos_agent
        #       Use the given directory for creating Unix domain sockets
        #   - log_file: <str>, e.g. /tmp/daos_agent.log
        #       Full path and name of the DAOS agent logfile.
        #   - control_log_mask: <str>, one of: error, info, debug
        #       Specifies the log level for agent logs.
        self.runtime_dir = BasicParameter(None, "/var/run/daos_agent")
        self.log_file = LogParameter(log_dir, None, "daos_agent.log")
        self.control_log_mask = BasicParameter(None, "debug")
        self.fabric_ifaces = []

    def update_log_file(self, name):
        """Update the log file name for the daos agent.

        If the log file name is set to None the log file parameter value will
        not be updated.

        Args:
            name (str): log file name
        """
        if name is not None:
            self.log_file.update(name, "agent_config.log_file")

class PerEngineYamlParameters(YamlParameters):
        """Defines the configuration yaml parameters for a single server."""

        # Engine environment variables that are required by provider type.
        REQUIRED_ENV_VARS = {
            "common": [
                "D_LOG_FILE_APPEND_PID=1",
                "COVFILE=/tmp/test.cov"],
            "ofi+tcp": [
                "SWIM_PING_TIMEOUT=10"],
            "ofi+verbs": [
                "FI_OFI_RXM_USE_SRX=1"],
            "ofi+cxi": [
                "FI_OFI_RXM_USE_SRX=1",
                "CRT_MRC_ENABLE=1"],
        }

        def __init__(self, index=None, provider=None):
            """Create a SingleServerConfig object.

            Args:
                index (int, optional): index number for the namespace path used
                    when specifying multiple servers per host. Defaults to None.
            """
            namespace = "/run/server_config/servers/*"
            if isinstance(index, int):
                namespace = "/run/server_config/servers/{}/*".format(index)
            super().__init__(namespace)
            if provider is not None:
                self._provider = provider
            else:
                self._provider = os.environ.get("CRT_PHY_ADDR_STR", "ofi+tcp")

            # Use environment variables to get default parameters
            default_interface = os.environ.get("DAOS_TEST_FABRIC_IFACE", "eth0")
            default_port = int(os.environ.get("OFI_PORT", 31416))
            default_share_addr = int(os.environ.get("CRT_CTX_SHARE_ADDR", 0))

            # All log files should be placed in the same directory on each host
            # to enable easy log file archiving by launch.py
            log_dir = os.environ.get("DAOS_TEST_LOG_DIR", "/tmp")

            # Parameters
            #   targets:                I/O service threads per engine
            #   first_core:             starting index for targets
            #   nr_xs_helpers:          I/O offload threads per engine
            #   fabric_iface:           map to OFI_INTERFACE=eth0
            #   fabric_iface_port:      map to OFI_PORT=31416
            #   log_mask:               map to D_LOG_MASK env
            #   log_file:               map to D_LOG_FILE env
            #   env_vars:               influences DAOS I/O Engine behavior
            #       Add to enable scalable endpoint:
            #           - CRT_CTX_SHARE_ADDR=1
            #           - CRT_CTX_NUM=8
            self.targets = BasicParameter(None, 8)
