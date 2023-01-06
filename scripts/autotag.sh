#!/bin/bash
RAWTAG=$(git describe --abbrev=0 --tags)
CURTAG=${RAWTAG##*/}
CURTAG="${CURTAG/v/}"

IFS='.' read -a vers <<<"$CURTAG"

MAJ=${vers[0]}
MIN=${vers[1]}
PATCH=${vers[2]}
echo "== INFO: Current Tag: $MAJ.$MIN.$PATCH"

case "$DRONE_COMMIT_MESSAGE" in
*"#major"*)
	((MAJ += 1))
	MIN=0
	PATCH=0
	echo "== INFO: #major found, incrementing major version"
	;;
*"#minor"*)
	((MIN += 1))
	PATCH=0
	echo "== INFO: #minor found, incrementing minor version"
	;;
*)
	((PATCH += 1))
	echo "== INFO: incrementing Patch Version"
	;;
esac

NEWTAG="$MAJ.$MIN.$PATCH"

GOSUMDB=off go work sync
git add .

FILES=$(find . -name go.mod | grep -v '^\./go.mod$')

for f in $FILES; do
  # --- Go Mod Tidy ---
	PACKAGE=$(echo $f | cut -d/ -f2- | xargs dirname)
  echo "== INFO: go mod tidy $PACKAGE"
	cd $PACKAGE
	GOSUMDB=off go mod tidy
	git add go.mod go.sum || true
	echo "== INFO: finished go mod tidy $PACKAGE"
	cd -
	# --- Update Tag in go.mod ---
	echo "== INFO: Updating $PACKAGE to $NEWTAG"
	sed -i -r "s#(\s+github\.com/tigorlazuardi/tower.*\sv).*#\1$NEWTAG#" $f
	echo "== INFO: finished updating $PACKAGE to $NEWTAG"
	git add $f
done

git commit -m "Bump Version to v$NEWTAG [CI SKIP]"

for f in $FILES; do
	PACKAGE=$(echo $f | cut -d/ -f2- | xargs dirname)
	PACKAGE_TAG="$PACKAGE/v$NEWTAG"
	echo "== INFO: Adding Tag: $PACKAGE_TAG"
	git tag $PACKAGE_TAG
done

echo "== INFO: Adding Tag: v$NEWTAG"
git tag v$NEWTAG

git push --force origin main
git push --tags
