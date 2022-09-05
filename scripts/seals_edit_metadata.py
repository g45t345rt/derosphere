import csv
import json
import sys

metadata_path = sys.argv[1]
rarity_path = sys.argv[2]
out_path = sys.argv[3]


def get_rarity():
    rarity_file = open(rarity_path, newline="")
    csv_rarity = csv.reader(rarity_file, delimiter=",")
    rarity = {}
    for row in csv_rarity:
        fileNumber = row[0]
        rarityValue = row[1]
        rarity[fileNumber] = rarityValue
    return rarity


def get_rarity_type(rarity):
    if (rarity >= 353.62):
        return "ultra_rare"
    elif (rarity >= 300.11):
        return "rare"
    else:
        return "common"


def main():
    metadata_file = open(metadata_path, "rb")
    data = json.load(metadata_file)
    rarity = get_rarity()
    nfts = []
    for cI in range(len(data["collection"])):
        collection = data["collection"][cI]
        id = collection["name"].replace("#", "")

        nfts.append({})
        nft = nfts[cI]
        nft["id"] = int(id)
        nft["rarity"] = float(rarity[id])
        nft["rarity_type"] = get_rarity_type(nft["rarity"])
        nft["attributes"] = {}

        attributes = collection["attributes"]
        for attribute in attributes:
            attr_type = attribute["trait_type"].lower().replace(" ", "_")
            attr_value = attribute["value"]
            attr_value = attr_value.replace("Untitled_Artwork ", "")
            nft["attributes"][attr_type] = attr_value

    metadata_file.close
    nfts_file = open(out_path, "w")
    json.dump(nfts, nfts_file, indent=2)
    nfts_file.close()


if __name__ == "__main__":
    main()
