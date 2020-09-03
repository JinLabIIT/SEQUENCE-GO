import pandas as pd

def load_data(number_thread, data_loc='./data/'):
    df_list = []
    for i in range(1, number_thread+1):
        try:
            temp_df = pd.read_csv(
                '{}{}/{}.txt'.format(data_loc, number_thread, i),
                header=None
            )
            temp_df.columns = ['time', 'event_number']
            temp_df.insert(2, 'thread_number', number_thread)
            # index start from 1
            temp_df.insert(3, 'thread_index', i)
            df_list.append(temp_df)
        except:
            pass
    return pd.concat(df_list, ignore_index=True)