# Password-Based Key Derivation Function 2
import hashlib

def handler(params):
    data = params.get("data", "user_secret_payload").encode("utf-8")
    rounds = int(params.get("rounds", 500_000))

    derived_key = hashlib.pbkdf2_hmac("sha256", data, b"static_salt", rounds)

    return {
        "key": derived_key.hex()
    }