import streamlit as st
from lib import get_response
import os
import logging
logging.basicConfig(level=logging.INFO)

st.set_page_config(page_title="AWS Cost and Usage Chatbot", page_icon="chart_with_upwards_trend", layout="centered", initial_sidebar_state="auto",
menu_items={
        'Get Help': 'https://docs.aws.amazon.com/cur/latest/userguide/cur-create.html',
        #'Report a bug':,
        'About': "# The purpose of this app is to help you get better understanding of your AWS Cost and Usage report!"
    })#HTML title
st.title("_:orange[Simplify] CUR data_ :sunglasses:")

def format_result(result):
    logging.info(result)
    logging.info("----")
    parts = result.split("\nSQLResult: ")
    if len(parts) > 1:
        sql_query = parts[0].replace("SQLQuery: ", "")
        sql_result = parts[1].strip("[]").split("), (")
        formatted_result = []
        for row in sql_result:
            formatted_result.append(tuple(item.strip("(),'") for item in row.split(", ")))
        return sql_query, formatted_result
    else:
        return result, []

def main():
    # Get the current directory
    current_dir = os.path.dirname(os.path.abspath(__file__))
    st.markdown("<div class='main'>", unsafe_allow_html=True)
    st.title("AWS Cost and Usage chatbot")
    st.write("Ask a question about your AWS Cost and Usage Report:")

    # Create a session state variable to store the chat history
    if "chat_history" not in st.session_state:
        st.session_state.chat_history = []

    user_input = st.text_input("You:", key="user_input")

    if user_input:
        try:
            result = get_response(user_input)
            sql_query, sql_result = format_result(result)
            st.write(sql_query)
            st.write(sql_result)
            st.code(sql_query, language="sql")
            if sql_result:
                st.write("SQLResult:")
                st.table(sql_result)
            else:
                st.write(result)
            st.session_state.chat_history.append({"user": user_input, "bot": result})
            st.text_area("Conversation:", value="\n".join([f"You: {chat['user']}\nBot: {chat['bot']}" for chat in st.session_state.chat_history]), height=300)
        except Exception as e:
            st.error(str(e))

    st.markdown("</div>", unsafe_allow_html=True)

if __name__ == "__main__":
    main()