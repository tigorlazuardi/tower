#!/bin/bash
CURTAG=$(git describe --abbrev=0 --tags)
CURTAG="${CURTAG/v/}"

IFS='.' read -a vers <<<"$CURTAG"

MAJ=${vers[0]}
MIN=${vers[1]}
PATCH=${vers[2]}
echo "Current Tag: $MAJ.$MIN.$PATCH"

case "$DRONE_COMMIT_MESSAGE" in
*"#major"*)
	((MAJ += 1))
	MIN=0
	PATCH=0
	echo "#major found, incrementing major version"
	;;
*"#minor"*)
	((MIN += 1))
	PATCH=0
	echo "#minor found, incrementing minor version"
	;;
*)
	((PATCH += 1))
	echo "incrementing Patch Version"
	;;
esac

NEWTAG="$MAJ.$MIN.$PATCH"
echo "Adding Tag: $NEWTAG"
git tag $NEWTAG

FILES=$(find . -name go.mod | grep -v '^\./go.mod$')

for f in $FILES; do
	sed -i "s#\(\s\+github\.com/tigorlazuardi/tower\sv\).\+#\1$NEWTAG#" $f
	git add $f
done
