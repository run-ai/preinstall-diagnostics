function install() {
    sed "s#_IMAGE_#${IMAGE}#; s#_NAMESPACE_#${NAMESPACE}#" res/resources.yaml | \
        kubectl apply -f -
    
    kubectl -n ${NAMESPACE} delete pod -l ${NAMESPACE}
}

function cleanup() {
    sed "s#_IMAGE_#${IMAGE}#; s#_NAMESPACE_#${NAMESPACE}#" res/resources.yaml | \
        kubectl -n ${NAMESPACE} delete -f - 
}

function show_logs() {
    POD=$(kubectl -n ${NAMESPACE} get pods | tail -n1 | cut -d" " -f1)
    kubectl -n ${NAMESPACE} wait --for=condition=Ready pod ${POD}
    kubectl -n ${NAMESPACE} logs ${POD} -f
}