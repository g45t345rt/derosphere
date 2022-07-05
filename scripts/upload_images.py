import os
import requests
import sys

account_id = sys.argv[1]
api_token = sys.argv[2]
folder_path = sys.argv[3]
prefix = sys.argv[4]


def upload_image(file, fileIndex):
    url = 'https://api.cloudflare.com/client/v4/accounts/' + \
        account_id + '/images/v1'
    headers = {'Authorization': 'Bearer ' + api_token}
    response = requests.post(url, headers=headers, files={
                             "id": prefix + fileIndex, "file": file})
    return response


def main():
    for filename in os.listdir(folder_path):
        f = os.path.join(folder_path, filename)
        if os.path.isfile(f):
            file = open(f, "rb")
            fileIndex = filename.split(".")[0]  # 1.webp
            res = upload_image(file, fileIndex)
            if res.status_code != 200:
                print(res.status_code, res.text)
            else:
                print(res.status_code, filename + " uploaded")


if __name__ == "__main__":
    main()
