set -eou pipefail

if [ "$1" = "log-exploration-oc-plugin" ]; then
	exec log-exploration-oc-plugin 
fi

exec "$@"