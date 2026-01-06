using Amazon.Lambda.Core;

[assembly: LambdaSerializer(typeof(Amazon.Lambda.Serialization.SystemTextJson.DefaultLambdaJsonSerializer))]

namespace dunno;

public class Function
{
    public string FunctionHandler(ILambdaContext context)
    {
        return context.LogStreamName;
    }
}
