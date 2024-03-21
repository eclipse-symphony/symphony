# Read the file in parameter and fill the array named "array"
getArray() {
    array=() # Create array
    while IFS= read -r line # Read a line
    do
        array+=("$line") # Append line to the array
    done < "$1"
}

getArray "done"

DEST_PATH=~/github/msftcoderdjw/staging-eclipse-symphony/api
SRC_PATH=~/github/msftcoderdjw/eclipse-symphony/api

for e in "${array[@]}"
do
    echo "Copying $e"
    cp $SRC_PATH/$e $DEST_PATH/$e
done

# cp -rf /home/jiadu/github/msftcoderdjw/eclipse-symphony/api/pkg/apis/v1alpha1/managers/* api/pkg/apis/v1alpha1/managers/
# cp -rf /home/jiadu/github/msftcoderdjw/eclipse-symphony/api/pkg/apis/v1alpha1/models/* api/pkg/apis/v1alpha1/models/
# cp -rf /home/jiadu/github/msftcoderdjw/eclipse-symphony/api/pkg/apis/v1alpha1/providers/* api/pkg/apis/v1alpha1/providers/
# cp -rf /home/jiadu/github/msftcoderdjw/eclipse-symphony/api/pkg/apis/v1alpha1/vendors/* api/pkg/apis/v1alpha1/vendors/