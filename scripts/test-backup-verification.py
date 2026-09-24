#!/usr/bin/env python3
"""Offline regression checks for offsite verification; never contacts R2."""
import hashlib
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
FAKE_RCLONE = """#!/usr/bin/env python3
import os, pathlib, sys
cmd = sys.argv[1]
remote = pathlib.Path(os.environ["TEST_REMOTE_FILE"])
if cmd == "listremotes":
    print("r2:")
elif cmd == "lsd":
    pass
elif cmd == "copy":
    assert "--immutable" in sys.argv
    assert "--checksum" in sys.argv
    if os.environ.get("TEST_COPY_FAIL"):
        sys.exit(1)
elif cmd == "size":
    print('{"count":1,"bytes":%d}' % remote.stat().st_size)
elif cmd == "cat":
    sys.stdout.buffer.write(remote.read_bytes())
else:
    sys.exit("unexpected rclone call: " + cmd)
"""


def main():
    (ROOT / "scratch").mkdir(exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="backup-verification-", dir=ROOT / "scratch") as directory:
        work = Path(directory)
        bindir = work / "bin"
        backups = work / "backups"
        bindir.mkdir()
        backups.mkdir()
        fake = bindir / "rclone"
        fake.write_text(FAKE_RCLONE)
        fake.chmod(0o700)
        local = backups / "prodcal-20260924-030001.sqlite3.gz"
        local.write_bytes(b"expected snapshot")
        remote = work / "remote"
        env = dict(os.environ, PATH=f"{bindir}:{os.environ['PATH']}",
                   BACKUP_DIR=str(backups), R2_REMOTE="r2",
                   R2_BUCKET="test-only", R2_PREFIX="db",
                   TEST_REMOTE_FILE=str(remote))

        def run():
            return subprocess.run(["bash", str(ROOT / "scripts/sync-to-r2.sh")],
                                  env=env, capture_output=True, text=True)

        remote.write_bytes(b"different content")
        assert remote.stat().st_size == local.stat().st_size
        result = run()
        assert result.returncode != 0, result.stdout
        assert "SHA-256 differs" in (backups / ".LAST-R2-FAILURE").read_text()
        assert not (backups / ".LAST-R2-SUCCESS").exists()

        remote.write_bytes(local.read_bytes())
        result = run()
        assert result.returncode == 0, result.stdout + result.stderr
        expected_hash = hashlib.sha256(local.read_bytes()).hexdigest()
        assert expected_hash in (backups / ".LAST-R2-SUCCESS").read_text()
        assert not (backups / ".LAST-R2-FAILURE").exists()

        env["TEST_COPY_FAIL"] = "1"
        result = run()
        assert result.returncode != 0
        assert (backups / ".LAST-R2-FAILURE").exists()
        assert remote.read_bytes() == local.read_bytes()
    print("PASS: mismatched bytes rejected; matching bytes verified; copy failure stays unhealthy")


if __name__ == "__main__":
    main()
