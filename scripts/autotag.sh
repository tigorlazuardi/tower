#!/bin/bash
RAWTAG=$(git describe --abbrev=0 --tags)
CURTAG=${RAWTAG##*/}
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

FILES=$(find . -name go.mod | grep -v '^\./go.mod$')

for f in $FILES; do
	sed -i -r "s#(\s+github\.com/tigorlazuardi/tower.*\sv).*#\1$NEWTAG#" $f
	git add $f
done

git commit -m "Bump Version to v$NEWTAG [CI SKIP]"

for f in $FILES; do
	PACKAGE=$(echo $f | cut -d/ -f2)
	PACKAGE_TAG="$PACKAGE/v$NEWTAG"
	echo "Adding Tag: $PACKAGE_TAG"
	git tag $PACKAGE_TAG
done

echo "Adding Tag: $NEWTAG"
git tag v$NEWTAG

git push --force origin main
git push --tags
