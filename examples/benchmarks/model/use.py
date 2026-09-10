import tensorflow as tf
import os


def handler(params, context):
    # Create the path to the weights file
    base_dir = os.path.dirname(os.path.abspath(__file__))
    weights_path = os.path.join(base_dir, 'my.weights.h5')

    # Recreate the model without weight
    model = tf.keras.applications.VGG16(weights=None, input_shape=(224, 224, 3))
    #model = tf.keras.applications.MobileNetV2(weights=None, input_shape=(224, 224, 3))

    # 2.Load the weights
    model.load_weights(weights_path)
    print("Weights loaded")

    # 3. Execute inference
    dummy_input = tf.random.normal([1, 224, 224, 3])
    predictions = model(dummy_input)

    return {'Inference done': "ok"}
