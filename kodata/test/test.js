class ToHostTranslator {
    constructor(clusterName) {
        this.ClusterName = clusterName;
    }

    translateName(namespace, name) {
        // we need to come up with a name which is:
        // - somewhat connectable to the original resource
        // - a valid k8s name
        // - idempotently calculatable
        // - unique for this combination of name/namespace/cluster
        const namePrefix = `${name}-${namespace}-${this.ClusterName}`;
        // use + as a separator since it can't be in an object name
        const nameKey = `${name}+${namespace}+${this.ClusterName}`;
        // it's possible that the suffix will be in the name, so we use hex to make it valid for k8s
        const nameSuffix = Buffer.from(nameKey).toString('hex');

        return this.safeConcatName(namePrefix, nameSuffix);
    }

    safeConcatName(...names) {
        const fullPath = names.join('-');
        if (fullPath.length < 64) {
            return fullPath;
        }

        const crypto = require('crypto');
        const digest = crypto.createHash('sha256').update(fullPath).digest('hex');

        // since we cut the string in the middle, the last char may not be compatible with what is expected in k8s
        // we are checking and if necessary removing the last char
        const c = fullPath.charCodeAt(56);
        if (('a'.charCodeAt(0) <= c && c <= 'z'.charCodeAt(0)) ||
            ('0'.charCodeAt(0) <= c && c <= '9'.charCodeAt(0))) {
            return fullPath.substring(0, 57) + "-" + digest.substring(0, 5);
        }

        return fullPath.substring(0, 56) + "-" + digest.substring(0, 6);
    }
}

let test = new ToHostTranslator('test1');

console.log(test.translateName('default', 'default-volume'))