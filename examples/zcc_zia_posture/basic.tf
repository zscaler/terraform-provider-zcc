resource "zcc_zia_posture" "this" {
  name     = "BD-Posture-Dev100"
  platform = "macos"

  high_trust_criteria = {
    cs = [
      {
        cn = [
          { id = "criterion-id-1", name = "Disk Encryption" },
          { id = "criterion-id-2", name = "Firewall Enabled" },
        ]
      }
    ]
  }

  medium_trust_criteria = {
    cs = [
      {
        cn = [
          { id = "criterion-id-3", name = "OS Up-To-Date" },
        ]
      }
    ]
  }

  low_trust_criteria = {
    cs = []
  }
}
