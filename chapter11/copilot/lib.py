from langchain_experimental.sql import SQLDatabaseChain
from langchain_community.utilities import SQLDatabase
from sqlalchemy import create_engine, URL
from langchain_aws import ChatBedrock as BedrockChat
from pyathena.sqlalchemy.rest import AthenaRestDialect
from langchain.chains import LLMChain
from langchain.prompts import PromptTemplate


class CustomAthenaRestDialect(AthenaRestDialect):
    def import_dbapi(self):
        import pyathena
        return pyathena

# DB Variables
connathena = "athena.eu-central-1.amazonaws.com"
portathena = '443'
schemaathena = 'cur'
s3stagingathena = 's3://mlabouardy-billing-dummy/cur-flattened-cur/'
wkgrpathena = 'primary'

# Build SQLAlchemy URL
url = URL.create(
    drivername="awsathena+rest",
    username="",  # anonymous connection string
    host=connathena,
    port=portathena,
    database=schemaathena,
    query={"s3_staging_dir": s3stagingathena, "work_group": wkgrpathena}
)

# Connect to Athena
engine_athena = create_engine(url, dialect=CustomAthenaRestDialect(), echo=False)
db = SQLDatabase(engine_athena)

# Setup LLM
model_kwargs = {"temperature": 0, "top_k": 250, "top_p": 1, "stop_sequences": ["\n\nHuman:"]}
llm = BedrockChat(model_id="anthropic.claude-3-5-sonnet-20240620-v1:0", model_kwargs=model_kwargs)

# Create the prompt
QUERY = PromptTemplate.from_template("""
Write only a valid Athena SQL query (no explanation) using the `datapoints20250703t015935z` table in the `cur` database.
Columns are in CUR v2 format, so wrap column names in double quotes (e.g., "lineitem/unblendedcost").
Do not alias or rename columns.
Return only the SQL query.

Question: {question}
""")
#db_chain = SQLDatabaseChain.from_llm(llm, db, verbose=True)

llm_chain = LLMChain(prompt=QUERY, llm=llm)

def get_response(user_input):
    sql_query = llm_chain.run(user_input).strip()

    if not sql_query.lower().startswith(("select", "with", "explain")):
        raise ValueError(f"Expected SQL query but got:\n{sql_query}")

    print("✅ Generated SQL Query:\n", sql_query)

    try:
        rows = db.run(sql_query)
    except Exception as e:
        return f"❌ SQLQuery:\n{sql_query}\n\nError: {e}"

    return f"SQLQuery: {sql_query}\nSQLResult: {rows}"
