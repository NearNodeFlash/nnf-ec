#!/bin/bash

# --- Function Definitions ---

# Prints a formal usage message
function print_usage_message {
  echo -e "Usage:\tversion.sh [-cbht] [-r id | -s version]\n"
  echo -e "Manages the semantic version of the repository.\n"
  echo -e "Options:"
  echo -e "  -b, --build\t\t\tAdd or update the build version to the current commit hash."
  echo -e "  -c, --clean\t\t\tRemove the build version, leaving only the release version."
  echo -e "  -h, --help \t\t\tDisplay this usage message."
  echo -e "  -t, --tag  \t\t\tTag this revision and push git tag upstream."
  echo -e "  -r, --release\tidentifier\tIncrement a part of the release version. Identifier must be one of \"major\", \"minor\", or \"patch\"."
  echo -e "  -s, --set\tversion\t\tManually set the version for the project. Version must follow semantic versioning rules and"
  echo -e "\t\t\t\tbe in the format <major>.<minor>.<patch>(+|-)<build_id>, e.g. 1.0.17+hj0laj9\n"
}

# Checks the semantic validity of a version by running the regex expression against it.
# Correct examples:
#   "0.0.1", "1.0.13+abcd123", "3.17.2-12abcd3" (if the build delimiter is '-')
function is_valid_version {
  [[ $(echo "$1" | perl -ne "print if /^(\d+\.){2}\d+([+-][a-z0-9]{7})?$/" | wc -l | xargs) -eq "1" ]] && true || false
}

# Manually sets the entire version, using the argument specified.
function set_version {

  if is_valid_version $1; then
    echo "Manually setting version to $1"
    echo "$1" > ${VERSION_FILE}
  else
    echo -e "\xE2\x9D\x8C $1 is not a valid version. Must be of the form <major>.<minor>.<patch>${BUILD_DELIM}<build>"
    exit 1
  fi
}

# Updates the build version to the current commit hash.
function update_build_version {

  echo "${RELEASE}${BUILD_DELIM}${CURRENT_HASH}" > ${VERSION_FILE}
  echo -e "\xE2\x9C\x85 Updated version from ${VERSION} to $(cat ${VERSION_FILE})"
}

# Removes the build version, leaving only the release version.
function clear_build_version {
  
  if [[ "${BUILD}" == "" ]]; then
    echo -e "\xE2\x9C\x85 Version \"${VERSION}\" already lacks a build version, no need to clear"
  else
    echo "${RELEASE}" > ${VERSION_FILE}
    echo -e "\xE2\x9C\x85 Updated version from ${VERSION} to $(cat ${VERSION_FILE})"
  fi
}

# Increments the patch version by 1.
function increment_patch {

  # Take off the build version and increment patch
  arr=(${RELEASE//./ })
  new_version="${arr[0]}.${arr[1]}.$((arr[2]+1))"

  # If there was a build version, add it back
  [[ "$BUILD" != "" ]] && new_version="${new_version}${BUILD_DELIM}${BUILD}"
  echo "${new_version}" > ${VERSION_FILE}
  echo -e "\xE2\x9C\x85 Incremented minor version from ${VERSION} to $(cat ${VERSION_FILE})"
}

# Increments the minor version by 1.
function increment_minor {

  # Take off the build version and increment minor 
  arr=(${RELEASE//./ })
  new_version="${arr[0]}.$((arr[1]+1)).${arr[2]}"

  # If there was a build version, add it back
  [[ "$BUILD" != "" ]] && new_version="${new_version}${BUILD_DELIM}${BUILD}"
  echo "${new_version}" > ${VERSION_FILE}
  echo -e "\xE2\x9C\x85 Incremented minor version from ${VERSION} to $(cat ${VERSION_FILE})"
}

# Increments the major version by 1.
function increment_major {

  # Take off the build version and increment major 
  arr=(${RELEASE//./ })
  new_version="$((arr[0]+1)).${arr[1]}.${arr[2]}"

  # If there was a build version, add it back
  [[ "$BUILD" != "" ]] && new_version="${new_version}${BUILD_DELIM}${BUILD}"
  echo "${new_version}" > ${VERSION_FILE}
  echo -e "\xE2\x9C\x85 Incremented minor version from ${VERSION} to $(cat ${VERSION_FILE})"
}

# Selects which part of the release version to update based on the argument provided.
function update_release_version {
 
  # Update based on requested identifier
  case $1 in
    major)
      increment_major
      ;;
    minor)
      increment_minor
      ;;
    patch)
      increment_patch
      ;;
    *)
      echo -e "\xE2\x9D\x8C Must specify \"major\", \"minor\", or \"patch\" for release.\nE.g.) -r minor OR --release minor\n"
      echo -e "  (You specified \"$1\".)"
      print_usage_message
      exit 1
  esac
}

# Uses the version found in the .version file to update the Helm SDP Manifest Config file.
function update_helm_sdp_manifest_config {

  sdp_manifest="helmSDPManifestConfig.yaml"
  new_version=$(cat ${VERSION_FILE})

  cat $sdp_manifest |  perl -pe "s/${PROJECT}_.+metadata\.json/${PROJECT}_${new_version}_metadata.json/g" > $sdp_manifest.tmp
  mv $sdp_manifest.tmp $sdp_manifest
  echo -e "\xE2\x9C\x85 Updated ${sdp_manifest} file to use \"${PROJECT}_${new_version}_metadata.json\""
}

# Uses the version found in the .version file to update the Helm Chart.yaml file.
function update_helm_chart_version {

  helm_chart="kubernetes/${PROJECT}/Chart.yaml"
  new_version=$(cat ${VERSION_FILE})

  cat $helm_chart |  perl -pe "s/^version: .+$/version: ${new_version}/g" > $helm_chart.tmp
  mv $helm_chart.tmp $helm_chart
  echo -e "\xE2\x9C\x85 Updated ${helm_chart} file to use \"version: ${new_version}\""
}

# Commits changes from updating the versions in files to the repo.
function commit_changes {

  new_version=$(cat ${VERSION_FILE})
  commit_msg="$(git rev-parse --abbrev-ref HEAD | cut -d '/' -f 2)

* Release version ${new_version}"
  if git add -A . && git commit -m "${commit_msg}" && git push; then
    echo -e "\xE2\x9C\x85 Successfully committed and pushed changes to version files"
  else
    echo -e "\xE2\x9D\x8C Unable to commit changes to version files"
  fi
}

# Tags the repository based off of the .version file.
function update_git_tag {
  
  old_tag=$(git describe --tags)
  new_version=$(cat ${VERSION_FILE})
  if git tag "v${new_version}" && git push origin "v${new_version}" ; then
    echo -e "\xE2\x9C\x85 Updated git tag from \"${old_tag}\" to \"v${new_version}\""
  else
    echo -e "\xE2\x9D\x8C Unable to push git tag upstream"
  fi
}

# Uses the now-updated .version file to sync version across SDP manifest
# file and Chart.yaml file, then commit those changes to the current branch.
function sync_version_changes {

  if [[ -d "./kubernetes" ]]; then
    update_helm_sdp_manifest_config
    update_helm_chart_version
  fi
  commit_changes 
}


# --- Global Variable Declarations ---
PROJECT=$(basename `git rev-parse --show-toplevel`)
PROJECT_ROOT=$(git rev-parse --show-toplevel) # Gives us the full path to the project root
CURRENT_HASH=$(git rev-parse --short HEAD)    # Gives us the shortened commit hash of HEAD
VERSION_FILE=${PROJECT_ROOT}/.version
VERSION=$(cat ${VERSION_FILE})
BUILD_DELIM="+"
RELEASE=""
BUILD=""


# Validates current version, and sets fields based one what pattern is matched.
# If a build is present, that field is set. 
# If no patterns are matched, the script errors out.
if [[ $(echo "${VERSION}" | perl -ne "print if /^(\d+\.){2}\d+[+-][a-z0-9]{7}$/") != "" ]]; then
  RELEASE=$(echo ${VERSION} | cut -d ${BUILD_DELIM} -f 1)
  BUILD=$(echo ${VERSION} | cut -d ${BUILD_DELIM} -f 2)
elif [[ $(echo "${VERSION}" | perl -ne "print if /^(\d+\.){2}\d+$/") != "" ]]; then
  RELEASE=${VERSION}
else
  echo "Previous version ${VERSION} is not valid. Must be of the form <major>.<minor>.<patch>${BUILD_DELIM}<build>."
  echo "Please manually set the version using ./version.sh --set <version>"
  exit 1
fi

# Parse script options and arguments, and call appropriate functions.
if [[ $# -eq 0 ]]; then
  update_build_version # Default, no args operation
else
  while [ "$1" != "" ]; do
    case "$1" in
      -s | --set)
        shift
        set_version $1
        sync_version_changes
        ;;
      -r | --release)
        shift
        update_release_version $1
        sync_version_changes
        ;;
      -b | --build)
        update_build_version
        sync_version_changes
        ;;
      -c | --clean)
        clear_build_version
        sync_version_changes
        ;;
      -t | --tag)
        update_git_tag
        ;;
      -h | --help)
        print_usage_message
        exit 0
        ;;
      *)
        echo "Option \"$1\" not recognized."
        print_usage_message
        exit 1
        ;;
    esac
    shift
  done
fi
