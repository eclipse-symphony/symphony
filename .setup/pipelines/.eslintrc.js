module.exports = {
    "extends": "@azure-iot",
    rules: {
        "@typescript-eslint/ban-types": "off",
        "@typescript-eslint/member-ordering": [
            "error",
            {
                "default": [
                    "public-static-field",
                    "public-instance-method"
                ]
            }
        ],
        "@typescript-eslint/naming-convention": [
            "error",
            {
                "selector": "class",
                "format": ["PascalCase"]
            }
        ],
        "@typescript-eslint/no-extraneous-class": "off",
        "@typescript-eslint/no-for-in-array": "off",
        "@typescript-eslint/no-floating-promises": "off",
        "@typescript-eslint/no-parameter-properties": "off",
        "@typescript-eslint/tslint/config/no-inferred-empty-object-type": "off",
        "prefer-arrow/prefer-arrow-functions": "off",
        "import/no-unassigned-import": "off",
        "no-fallthrough": "off",
        "no-param-reassign": "off",
        "no-restricted-syntax": "off",
        "no-sparse-arrays": "off",
        "no-template-curly-in-string": "off",
        "no-throw-literal": "off",
        "no-useless-constructor": "off",
        "radix": "off",
        "unicorn/filename-case": "off"
    },
    ignorePatterns: []
};
