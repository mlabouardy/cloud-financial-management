### Azure Policy

Similar to AWS, Azure offers a service called [Azure Policy](https://learn.microsoft.com/en-us/azure/governance/policy/overview) that can enforce organizational tagging policies, ensuring all resources follow consistent tagging guidelines. Azure Policy can audit resources to verify proper tagging and can even apply missing tags automatically.

To create a tagging policy, go to the Azure Portal and search for the “Policy” service using the search bar.





![alt_text](../assets/images/chapter4/image1.png "image_tooltip")



###### Figure 4.24. Azure policy service page

Then, go to “Definitions” within the “Authoring” section, where you’ll find a list of predefined policies that you can use. For example, you can leverage the “Require a tag on resources” policy to prevent the launch of Azure resources without a predefined tag key. The policy configuration is shown in the screenshot:



![alt_text](../assets/images/chapter4/image2.png "image_tooltip")
 


###### Figure 4.25. Azure policy example

Click on the “Assign” button and set the subscription you want to target as the scope for this policy, leaving the rest as default. In the “Parameters” section, set *CostCenter* as the required tag key.

![alt_text](../assets/images/chapter4/image3.png "image_tooltip")



###### Figure 4.26. Making CostCenter tag key a requirement

On the next tab, leave the default values for the “Remediation” section and set a non-compliance message that will be displayed to the user when a resource is created without the *CostCenter* tag key.


![alt_text](../assets/images/chapter4/image4.png "image_tooltip")



###### Figure 4.27. Example of non-compliance message

Review and create the policy. On the policy information page, it will evaluate existing Azure resources and mark those missing the *CostCenter* tag key as non-compliant, as shown in the screenshot below.


![alt_text](../assets/images/chapter4/image5.png "image_tooltip")



###### Figure 4.28. List of Azure resources missing the CostCenter tag key

Let’s test it with new resources by attempting to launch an Azure VM without the *CostCenter* tag key. The policy will prevent the VM deployment, and a “Validation failed” error message will appear. Clicking on it will display the tagging policy non-compliance message we configured above, as shown in the screenshot.


![alt_text](../assets/images/chapter4/image6.png "image_tooltip")



###### Figure 4.29. Unable to launch a VM due to missing CostCenter tag

Now, we’ve a policy that ensures that every Azure resource within the specified scope has a *CostCenter* tag, reducing the chances of missing or incorrect tags across resources. As a result, it increases cost transparency and accountability within the organization.

You can take this further and repeat the same steps to create a “Required a tag on a resource group” policy as well as to enforce tagging at the resource group level:


![alt_text](../assets/images/chapter4/image7.png "image_tooltip")



###### Figure 4.30. Enforcing tagging at the resource group

While cloud-native services provide solid options for managing and enforcing tags, third-party open-source tools can offer additional flexibility and features, especially in multi-cloud environments. In the next section, we’ll explore some popular open-source tools that can help teams maintain consistent tagging across different cloud providers.
